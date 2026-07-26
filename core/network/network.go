package network

import (
	"fmt"

	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/packet"
)

// Network 网络服务
type Network struct {
	domain.IServer // 网络服务器
}

// NewNetwork 创建 Network（泛型工厂）
func NewNetwork[T any](f func() *T) *Network {
	return &Network{IServer: any(f()).(domain.IServer)}
}

// Unbind 解绑并移除连接
func (d *Network) Unbind(socketId uint32) {
	d.Del(socketId)
}

// SendToClient 发送消息到客户端
func (d *Network) SendToClient(head *packet.Head, body []byte) error {
	if client := d.Get(head.Uid); client != nil {
		return client.Send(head, body)
	}
	return fmt.Errorf("连接不存在, socketId:%d, cmd:%d", head.SocketId, head.Cmd)
}
