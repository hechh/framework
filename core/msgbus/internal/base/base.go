package base

import (
	"fmt"
	"github.com/hechh/framework/packet"
	"sync"
)

var (
	bytes = sync.Pool{
		New: func() any {
			return make([]byte, 0, 512)
		},
	}
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

// 构建当前节点的单播主题
func BuildSelfPoint(cfg *packet.NodeConfig) string {
	return BuildPoint(cfg.Type, cfg.Id)
}

// 构建当前节点的广播主题
func BuildSelfBroadcast(nodeCfg *packet.NodeConfig) string {
	return BuildBroadcast(nodeCfg.Type)
}

// 构建当前节点的回复主题
func BuildSelfReply(cfg *packet.NodeConfig) string {
	return BuildReply(cfg.Type, cfg.Id)
}
