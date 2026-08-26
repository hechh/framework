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

func New() *Client {
	return new(Client)
}

func (m *Client) Init(cfg *redispool.Config) error {
	m.Client = &goredis.Client{}
	s, err := miniredis.Run()
	if err != nil {
		return err
	}
	m.miniredis = s

	port, _ := strconv.Atoi(s.Port())
	return m.Client.Init(&redispool.Config{
		Ip:     s.Host(),
		Port:   uint32(port),
		Db:     0,
		Prefix: "mock",
	})
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
