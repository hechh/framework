package postgres

import (
	"fmt"
	"github.com/hechh/framework/pkg/dbpool/internal/base"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/dbpool"
	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

type Client struct {
	engine  *xorm.EngineGroup
	dsn     []string
	dbname  string
	tables  []any
	isAlive int32
	synced  bool
}

func NewClient() *Client {
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

	if d.engine != nil {
		_ = d.engine.Close()
	}

	d.engine = eng
	atomic.StoreInt32(&d.isAlive, 1)
	return nil
}

func (d *Client) Close() error {
	if d.engine != nil {
		atomic.StoreInt32(&d.isAlive, 0)
		return d.engine.Close()
	}
	return nil
}

func (d *Client) Ping() error {
	if d.engine == nil {
		return fmt.Errorf("database engine is nil")
	}
	return d.engine.Ping()
}

func (d *Client) IsAlive() bool {
	return atomic.LoadInt32(&d.isAlive) == 1
}

func (d *Client) Engine() *xorm.EngineGroup {
	return d.engine
}

func (d *Client) NewSession() *xorm.Session {
	if d.engine != nil {
		return d.engine.NewSession()
	}
	return nil
}
