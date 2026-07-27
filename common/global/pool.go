package global

import (
	"sync"

	"github.com/hechh/framework/packet"
)

var (
	headPool = sync.Pool{
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
