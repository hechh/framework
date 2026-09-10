package mockredis

import (
	"strconv"

	"github.com/alicebob/miniredis/v2"
	"github.com/hechh/framework/pkg/redispool"
	"github.com/hechh/framework/pkg/redispool/adapter/goredis"
)

type Client struct {
	*goredis.Client
	miniredis *miniredis.Miniredis
}

func New(cfg *redispool.Config) (*Client, error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	port, _ := strconv.Atoi(s.Port())
	// 必须透传 cfg.DbName 与 cfg.Prefix：RedisPool 以 cli.DbName() 为键注册客户端
	// （redispool.Get(name) / GetByHash），丢弃 DbName 会注册到空键上，业务按名称取不到客户端
	client, err := goredis.New(&redispool.Config{
		DbName: cfg.DbName,
		Ip:     s.Host(),
		Port:   uint32(port),
		Db:     0, // miniredis 仅支持 db 0
		Prefix: cfg.Prefix,
	})
	if err != nil {
		return nil, err
	}
	return &Client{Client: client, miniredis: s}, nil
}

func (m *Client) Close() error {
	if m.Client != nil {
		_ = m.Client.Close()
	}
	if m.miniredis != nil {
		m.miniredis.Close()
	}
	return nil
}
