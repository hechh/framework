package postgre

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/pkg/dbpool/internal/base"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/dbpool"
	_ "github.com/lib/pq"
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
			"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
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
	eng, err := xorm.NewEngineGroup("postgres", d.dsn)
	if err != nil {
		return fmt.Errorf("create engine error: %w", err)
	}

	base.SetupEngine(eng)

	if err := base.SyncTables(eng, d.tables, &d.synced); err != nil {
		_ = eng.Close()
		return fmt.Errorf("sync tables error: %w", err)
	}

	if err := eng.Ping(); err != nil {
		_ = eng.Close()
		return fmt.Errorf("ping error: %w", err)
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
