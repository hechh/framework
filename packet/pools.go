package packet

import (
	"sync"
)

var (
	headPool = sync.Pool{
		New: func() any { return new(Head) },
	}
)

func GetHead(opts ...func(*Head)) *Head {
	obj := headPool.Get().(*Head)
	for _, opt := range opts {
		opt(obj)
	}
	return obj
}

func PutHead(val *Head) {
	*val = Head{}
	headPool.Put(val)
}
