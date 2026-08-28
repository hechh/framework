package dbpool

import (
	"sync"
	"time"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/dbpool/internal/registry"
	"github.com/hechh/framework/pkg/mlog"
	"xorm.io/xorm"
)

// Config 数据库分片配置
type Config struct {
	DbType   string            `yaml:"db_type,omitempty"`  // 数据库类型，默认为mysql
	DbName   string            `yaml:"dbname,omitempty"`   // 数据库名称
	Db       uint32            `yaml:"db,omitempty"`       // 数据库编号
	User     string            `yaml:"user,omitempty"`     // 数据库用户名
	Password string            `yaml:"password,omitempty"` // 数据库密码
	Host     string            `yaml:"host,omitempty"`     // 数据库主机地址
	Port     uint32            `yaml:"port,omitempty"`     // 数据库端口
	Slaves   map[int32]*Config `yaml:"slaves,omitempty"`   // 从库配置列表
}

type IClient interface {
	Init(*Config, ...any) error
	Close() error
	Connect() error
	Ping() error
	IsAlive() bool
	Engine() *xorm.EngineGroup
	NewSession() *xorm.Session
}

type DbPool struct {
	newFunc   func(string) IClient // new函数
	pools     map[string]IClient   // 全局数据库连接
	exitCh    chan struct{}        // 关闭通道
	closeOnce sync.Once            // 确保 Close 只执行一次
}

func NewDbPool(f func(string) IClient) *DbPool {
	return &DbPool{
		newFunc: f,
		pools:   make(map[string]IClient),
		exitCh:  make(chan struct{}),
	}
}

// 初始化全局数据库
func (d *DbPool) Init(cfgs map[string]*Config) error {
	for name, dbcfg := range cfgs {
		cli := d.newFunc(dbcfg.DbType)
		if err := cli.Init(dbcfg, registry.GetTables(name)...); err != nil {
			d.Close()
			return err
		}
		d.pools[name] = cli
	}

	go d.check()
	return nil
}

func (d *DbPool) Close() {
	d.closeOnce.Do(func() {
		close(d.exitCh)
		for name, cli := range d.pools {
			if err := cli.Close(); err != nil {
				mlog.Errorf("关闭全局库[%s]失败: %v", name, err)
			}
		}
	})
}

func (d *DbPool) check() {
	tt := time.NewTicker(30 * time.Second)
	defer tt.Stop()
	for {
		select {
		case <-tt.C:
			for _, cli := range d.pools {
				d.checkClient(cli)
			}
		case <-d.exitCh:
			return
		}
	}
}

func (d *DbPool) checkClient(cli IClient) {
	if err := cli.Ping(); err != nil {
		mlog.Errorf("数据库连接异常断开: %v", err)
		if err := safe.Retry(3, 3*time.Second, cli.Connect); err != nil {
			mlog.Errorf("数据库重连全部失败, error:%v", err)
		} else {
			mlog.Infof("数据库重连成功")
		}
	}
}

func (d *DbPool) Get(name string) IClient {
	if val, ok := d.pools[name]; ok && val.IsAlive() {
		return val
	}
	return nil
}
