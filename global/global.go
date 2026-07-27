package global

import (
	"sync"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/utils"
)

var (
	CmdConvertor    utils.IConvertor
	NodeConvertor   utils.IConvertor
	Self            *packet.Node
	GatewayNodeType uint32
	headPool        = sync.Pool{
		New: func() any { return new(packet.Head) },
	}
)

func SetNodeConvertor(n map[string]int32, i map[int32]string) {
	NodeConvertor = utils.WrapConvertor(n, i)
}

func SetCmdConvertor(n map[string]int32, i map[int32]string) {
	CmdConvertor = utils.WrapConvertor(n, i)
}

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
