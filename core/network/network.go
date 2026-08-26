package network

import (
	"github.com/hechh/framework/packet"
)

// IServer 网络服务端接口
type INetwork interface {
	Init(string) error
	Close()
	Bind(uint32, uint64) bool
	Unbind(uint32, uint64)
	Send(*packet.Head, []byte) error
}

type Network struct {
	net     INetwork
	newFunc func() INetwork
}

func NewNetwork[T INetwork](f func() T) *Network {
	return &Network{
		newFunc: func() INetwork { return f() },
	}
}

func (d *Network) Init(addr string) error {
	d.net = d.newFunc()
	return d.net.Init(addr)
}

func (d *Network) Close() {
	d.net.Close()
}

func (d *Network) Bind(socketId uint32, uid uint64) bool {
	return d.net.Bind(socketId, uid)
}

func (d *Network) Unbind(socketId uint32, uid uint64) {
	d.net.Unbind(socketId, uid)
}

func (d *Network) Send(head *packet.Head, body []byte) error {
	return d.net.Send(head, body)
}
