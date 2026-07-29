package global

import (
	"sync"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/enum"
)

var (
	CmdConvertor    enum.IConvertor
	NodeConvertor   enum.IConvertor
	Self            *packet.Node
	GatewayNodeType uint32
	headPool        = sync.Pool{
		New: func() any { return new(packet.Head) },
	}
)

func GetHead(opts ...func(*packet.Head)) *packet.Head {
	obj := headPool.Get().(*packet.Head)
	for _, opt := range opts {
		opt(obj)
	}
	return obj
}

func PutHead(val *packet.Head) {
	*val = packet.Head{}
	headPool.Put(val)
}
