package global

import (
	"framework/packet"
	"sync/atomic"
)

var (
	self              *packet.Node
	actorIdGenerator  = uint64(0) // actorId生成器
	socketIdGenerator = uint32(0)
)

func SetSelf(nn *packet.Node) {
	self = nn
}

func GetSelf() *packet.Node {
	return self
}

func GetSelfType() uint32 {
	return self.Type
}

func GetSelfId() uint32 {
	return self.Id
}

// 生成 actorId
func GenerateActorId() uint64 {
	return atomic.AddUint64(&actorIdGenerator, 1)
}

func GenerateSocketId() uint32 {
	return atomic.AddUint32(&socketIdGenerator, 1)
}
