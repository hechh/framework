package mockdb

import (
	"fmt"
	"io"
	"sync/atomic"

	"github.com/hechh/framework/pkg/dbpool"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

var mockClientSeq atomic.Int64

type Client struct {
	name    string
	xGroup  *xorm.EngineGroup // SQLite 内存数据库引擎组
	pingFn  func() error
	closeFn func() error
	aliveFn func() bool
	isAlive int32
}

func NewClient(_ string) dbpool.IClient {
	return &Client{}
}

func (m *Client) Init(cfg *dbpool.Config, tables ...any) error {
	// 使用原子递增序号生成唯一 URI，确保每个 Client 实例拥有独立的数据库
	seq := mockClientSeq.Add(1)
	dsn := fmt.Sprintf("file:%s_%d?mode=memory&cache=shared", cfg.DbName, seq)
	eg, err := xorm.NewEngineGroup("sqlite", []string{dsn})
	if err != nil {
		return err
	}
	eg.SetLogger(log.NewSimpleLogger(io.Discard))
	eg.SetMaxIdleConns(1)
	eg.SetMaxOpenConns(1)

	// 初始化
	m.name = cfg.DbName
	m.xGroup = eg
	atomic.StoreInt32(&m.isAlive, 1)

	if len(tables) > 0 {
		return m.xGroup.Engine.Sync2(tables...)
	}
	return nil
}

func (m *Client) Connect() error {
	return nil
}

func (m *Client) Close() error {
	atomic.StoreInt32(&m.isAlive, 0)
	if m.closeFn != nil {
		return m.closeFn()
	}
	if m.xGroup != nil {
		return m.xGroup.Close()
	}
	return nil
}

func (m *Client) Ping() error {
	if m.pingFn != nil {
		return m.pingFn()
	}
	if m.xGroup != nil {
		return m.xGroup.Ping()
	}
	return nil
}

func (m *Client) IsAlive() bool {
	if m.aliveFn != nil {
		return m.aliveFn()
	}
	return atomic.LoadInt32(&m.isAlive) == 1
}

func (m *Client) Engine() *xorm.EngineGroup {
	return m.xGroup
}

func (m *Client) NewSession() *xorm.Session {
	if m.xGroup != nil {
		return m.xGroup.NewSession()
	}
	return nil
}
