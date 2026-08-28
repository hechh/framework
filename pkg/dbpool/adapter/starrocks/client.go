package starrocks

import (
	"fmt"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/dbpool"
	"github.com/hechh/framework/pkg/dbpool/internal/base"
	"xorm.io/xorm"
)

type Client struct {
	engine  atomic.Pointer[xorm.EngineGroup] // 原子读写，避免重连/关闭与业务并发访问的数据竞争
	dsn     []string
	dbname  string
	tables  []any
	isAlive int32
	synced  bool
}

func NewClient(_ string) dbpool.IClient {
	return &Client{}
}

func (d *Client) Init(cfg *dbpool.Config, tables ...any) error {
	d.dsn = append(d.dsn,
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?timeout=3s&readTimeout=10s&writeTimeout=10s&parseTime=true&charset=utf8mb4&loc=Local&tls=false&interpolateParams=true",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DbName,
		),
	)
	d.dbname = cfg.DbName
	d.tables = tables

	return safe.Retry(3, 2*time.Second, d.Connect)
}

func (d *Client) Connect() error {
	eng, err := xorm.NewEngineGroup("mysql", d.dsn)
	if err != nil {
		return err
	}

	base.SetupEngine(eng)

	if err := eng.Ping(); err != nil {
		_ = eng.Close()
		return err
	}

	// 原子替换旧引擎并关闭，避免“先读后写”的竞争窗口
	if old := d.engine.Swap(eng); old != nil {
		_ = old.Close()
	}

	atomic.StoreInt32(&d.isAlive, 1)
	return nil
}

func (d *Client) Close() error {
	if old := d.engine.Swap(nil); old != nil {
		atomic.StoreInt32(&d.isAlive, 0)
		return old.Close()
	}
	return nil
}

func (d *Client) Ping() error {
	eng := d.engine.Load()
	if eng == nil {
		return fmt.Errorf("database engine is nil")
	}
	return eng.Ping()
}

func (d *Client) IsAlive() bool {
	return atomic.LoadInt32(&d.isAlive) == 1
}

func (d *Client) Engine() *xorm.EngineGroup {
	return d.engine.Load()
}

func (d *Client) NewSession() *xorm.Session {
	eng := d.engine.Load()
	if eng != nil {
		return eng.NewSession()
	}
	return nil
}
