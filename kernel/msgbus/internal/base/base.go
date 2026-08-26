package base

import (
	"fmt"
	"sync"

	"github.com/hechh/framework/packet"
)

var (
	bytes = sync.Pool{
		New: func() any {
			return make([]byte, 0, 512)
		},
	}
	packetFunc func(*packet.Packet)
)

func GetBytes() []byte {
	return bytes.Get().([]byte)
}

func PutBytes(val []byte) {
	if cap(val) <= 32*1024 {
		val = val[:0]
		bytes.Put(val)
	}
}

func SetPacketFunc(f func(*packet.Packet)) {
	packetFunc = f
}

func PacketHandler(msg *packet.Packet) {
	if packetFunc != nil {
		packetFunc(msg)
	}
}

// 构建 NATS 单播主题
func BuildPoint(nodeType, nodeId uint32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

// 构建回复主题
func BuildReply(nodeType, nodeId uint32) string {
	return fmt.Sprintf("%d/%d/reply", nodeType, nodeId)
}

// 构建广播主题
func BuildBroadcast(nodeType uint32) string {
	return fmt.Sprintf("%d/broadcast", nodeType)
}
