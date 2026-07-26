package domain

import "github.com/hechh/framework/packet"

const (
	DEFAULT_VIRTUAL_COUNT = 100
)

type IDiscovery interface {
	Init(*packet.DiscoveryConfig) error // 初始化
	Close()                             // 关闭发现服务
	Register(string, []byte) error      // 注册节点
	Watch(func(string, []byte)) error   // 监听节点变化
}
