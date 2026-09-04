package client

import (
	"github.com/hechh/framework/pkg/redispool"
)

type Client struct {
	uid uint64
	redispool.IClient
	old redispool.IClient
}
