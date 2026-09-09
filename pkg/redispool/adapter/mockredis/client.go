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
	client, err := goredis.New(&redispool.Config{
		Ip:     s.Host(),
		Port:   uint32(port),
		Db:     0,
		Prefix: "mock",
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
