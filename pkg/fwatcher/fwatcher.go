package fwatcher

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/library/logic"
	"github.com/hechh/framework/library/safe"
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

func (d *FWatcher) save(path string, body []byte) {
	// 删除事件（body==nil）：清空等内部操作触发，忽略避免破坏本地文件
	if body == nil {
		return
	}
	sheet := strings.TrimPrefix(path, d.cfg.Etcd.Prefix+"/")
	filename := filepath.Join(d.abspath, sheet+d.cfg.Ext)

	// 落地保存：内容一致则跳过，否则（含文件不存在/其他读取错误）统一写入
	old, err := os.ReadFile(filename)
	if err == nil && bytes.Equal(old, body) {
		return
	}
	if err != nil && !os.IsNotExist(err) {
		mlog.Warnf("读取本地配置(%s)失败: %v", filename, err)
	}
	if err := fileutil.Save(filename, body); err != nil {
		mlog.Errorf("收到同步配置，但是保存失败 error=%v", err)
	}
}

func (d *FWatcher) Init(cfg *Config) error {
	// 确保数据目录存在（远程同步写文件与本地监听都依赖它）
	var err error
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
	if cfg.IsSync {
		if err := d.sync.Clear(); err != nil {
			return err
		}
	}
	// 建立远程配置变更监听
	if err := d.sync.Watch(d.save); err != nil {
		return err
	}

	// 获取所有变更配置
	files, err := registry.Glob(d.pattern)
	if err != nil {
		return err
	}

	// 同步配置：先清空 etcd 中所有 kv，再全量上传本地配置，保证 etcd 与本地一致
	if cfg.IsSync {
		for sheet, file := range files {
			if err := d.sync.Put(sheet, file.GetText()); err != nil {
				return err
			}
		}
	}

	// 建立监听目录
	if d.fswatcher, err = fsnotify.NewWatcher(); err != nil {
		return err
	}
	if err = d.fswatcher.Add(d.abspath); err != nil {
		return err
	}

	// 加载配置
	if err := registry.Load(files); err != nil {
		return err
	}
	// 监听本地目录文件变更
	safe.SafeGo(mlog.Fatalf, d.watch)
	return nil
}

func (d *FWatcher) Close() {
	close(d.exitCh)
	if d.sync != nil {
		d.sync.Close()
	}
}

func (d *FWatcher) watch() {
	defer d.fswatcher.Close()
	for {
		select {
		case <-d.exitCh:
			return
		case event, ok := <-d.fswatcher.Events:
			if !ok {
				return
			}
			if logic.Has(event.Op, fsnotify.Write) || logic.Has(event.Op, fsnotify.Create) {
				files, err := registry.Glob(d.pattern)
				if err != nil {
					mlog.Errorf("配置文件读取失败 abspath:%s, error=%v", d.abspath, err)
					continue
				}

				// 同步：先清空 etcd 中所有 kv，再全量上传本地配置
				if d.cfg.IsSync {
					for sheet, file := range files {
						if err := d.sync.Put(sheet, file.GetText()); err != nil {
							mlog.Errorf("同步游戏配置失败，sheet:%s, error:%v", sheet, err)
						}
					}
				}

				// 重新加载到内存
				if err := registry.Load(files); err != nil {
					mlog.Errorf("游戏配置加载失败 error=%v", err)
				}
			}
		}
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
