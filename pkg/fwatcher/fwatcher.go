package fwatcher

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/fwatcher/internal/parser"
	"github.com/hechh/framework/pkg/fwatcher/internal/registry"
	"github.com/hechh/framework/pkg/mlog"
)

type EtcdConfig struct {
	Prefix    string   `yaml:"prefix,omitempty"`     // etcd 前缀主题
	Endpoints []string `yaml:"endpoints,omitempty"`  // etcd 节点地址列表
	KeepAlive int64    `yaml:"keep_alive,omitempty"` // 保活时间（秒）
}

type Config struct {
	IsSync   bool        `yaml:"-"`                   // 是否开启配置同步
	DataPath string      `yaml:"data_path,omitempty"` // 数据文件目录
	XlsxPath string      `yaml:"xlsx_path,omitempty"` // Excel 文件目录
	Ext      string      `yaml:"ext,omitempty"`       // 文件扩展名
	Etcd     *EtcdConfig `yaml:"etcd,omitempty"`      // etcd 配置
}

// 配置同步接口
type ISync interface {
	Init(*Config) error
	Close()
	Clear() error
	Put(string, []byte) error
	Update(string, []byte) error
	Delete(string) error
	Watch(func(string, []byte)) error
}

// FWatcher 文件监听器
type FWatcher struct {
	newFunc   func() ISync
	pattern   string            // 匹配模式
	abspath   string            // 配置路径
	cfg       *Config           // 配置
	sync      ISync             // 远程同步
	fswatcher *fsnotify.Watcher // 监听
	exitCh    chan struct{}     // 退出
}

func NewFWatcher[T ISync](f func() T) *FWatcher {
	return &FWatcher{
		newFunc: func() ISync { return f() },
		exitCh:  make(chan struct{}),
	}
}

func (d *FWatcher) Init(cfg *Config) (err error) {
	d.cfg = cfg
	if d.abspath, err = filepath.Abs(cfg.DataPath); err != nil {
		return err
	}
	if err = fileutil.EnsureDir(d.abspath); err != nil {
		return err
	}
	d.pattern = fmt.Sprintf("%s/*%s", d.abspath, cfg.Ext)

	// 建立连接
	d.sync = d.newFunc()
	if err := d.sync.Init(cfg); err != nil {
		return err
	}

	// 获取所有变更配置
	files, err := registry.Glob(d.pattern)
	if err != nil {
		return err
	}

	if cfg.IsSync {
		// 先清空
		if err := d.sync.Clear(); err != nil {
			return err
		}
		// 同步配置：先清空 etcd 中所有 kv，再全量上传本地配置，保证 etcd 与本地一致
		for sheet, file := range files {
			if err := d.sync.Put(sheet, file); err != nil {
				return err
			}
		}
	}

	// 先加载本地配置
	for sheet, file := range files {
		if par := registry.GetParser(sheet); par != nil {
			if err := par.Parse(file); err != nil {
				return err
			}
		}
	}

	// 然后同步最新配置，并且监听
	if err := d.sync.Watch(d.save); err != nil {
		return err
	}

	// 检测是否全部加载
	return registry.WalkParser(func(sheet string, par parser.IParser) error {
		if !par.IsLoaded() {
			return fmt.Errorf("配置未加载，sheet:%s", sheet)
		}
		return nil
	})
}

func (d *FWatcher) save(path string, body []byte) {
	// 删除事件（body==nil）：清空等内部操作触发，忽略避免破坏本地文件
	sheet := strings.TrimPrefix(path, d.cfg.Etcd.Prefix+"/")
	if body == nil {
		mlog.Errorf("删除配置%s", sheet)
		return
	}

	// 原子保存文件
	filename := filepath.Join(d.abspath, sheet+d.cfg.Ext)
	if err := fileutil.AtomicSave(filename, body); err != nil {
		mlog.Errorf("变更配置%s保存失败: %v", sheet, err)
	} else {
		mlog.Infof("变更配置%s保存成功", sheet)
	}

	// 判断fileInfo是否存在
	info := registry.GetFileInfo(sheet)
	if info != nil && !info.IsChange(body) {
		return
	}
	if info == nil {
		info = parser.NewFileInfo(filename, body)
		registry.RegisterFileInfo(sheet, info)
	}

	// 直接加载配置
	if par := registry.GetParser(sheet); par != nil {
		if err := par.Parse(body); err != nil {
			mlog.Errorf("配置(%s)加载失败 error=%v", sheet, err)
		} else {
			// 更新md5值
			info.Update(body)
		}
	}
}

func (d *FWatcher) Close() {
	close(d.exitCh)
	if d.sync != nil {
		d.sync.Close()
	}
}

// Clear 清空 etcd 中所有配置（发布前清理残留）。
func (d *FWatcher) Clear() error {
	if d.sync == nil {
		return fmt.Errorf("配置同步服务未初始化")
	}
	return d.sync.Clear()
}

// Put 上传配置到 etcd。
func (d *FWatcher) Put(key string, msg []byte) error {
	if d.sync == nil {
		return fmt.Errorf("配置同步服务未初始化")
	}
	return d.sync.Put(key, msg)
}

// Update 更新配置到 etcd（key 不存在时报错）。
func (d *FWatcher) Update(key string, msg []byte) error {
	if d.sync == nil {
		return fmt.Errorf("配置同步服务未初始化")
	}
	return d.sync.Update(key, msg)
}

// Delete 删除 etcd 中的配置。
func (d *FWatcher) Delete(key string) error {
	if d.sync == nil {
		return fmt.Errorf("配置同步服务未初始化")
	}
	return d.sync.Delete(key)
}
