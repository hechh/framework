package migrate

import "github.com/hechh/framework/pkg/redispool"

type Client struct {
	oldCli redispool.IClient
	newCli redispool.IClient
}

func NewClient(n, o redispool.IClient) *Client {
	return &Client{oldCli: o, newCli: n}
}
