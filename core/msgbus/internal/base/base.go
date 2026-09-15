package base

import (
	"fmt"

	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

var (
	packetFunc func(*packet.Packet)
)

func SetPacketFunc(f func(*packet.Packet)) {
	packetFunc = f
}

func PacketHandler(msg *packet.Packet) {
	// 兜底：非法包（nil 或缺少 Head）在此丢弃，避免下游任意位置 nil 解引用
	if msg == nil || msg.Head == nil {
		mlog.Errorf("[nats] PacketHandler 收到无效消息, msg=%v", msg)
		return
	}
	if packetFunc != nil {
		packetFunc(msg)
	}
}

// 构建 NATS 单播主题
func BuildPoint(nodeType, nodeId uint32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

// BuildPointSlot 构建分片后的单播主题：{type}/{id}/{slot}
func BuildPointSlot(nodeType, nodeId uint32, slot uint32) string {
	return fmt.Sprintf("%d/%d/%d", nodeType, nodeId, slot)
}

// 构建回复主题
func BuildReply(nodeType, nodeId uint32) string {
	return fmt.Sprintf("%d/%d/reply", nodeType, nodeId)
}

// 构建广播主题
func BuildBroadcast(nodeType uint32) string {
	return fmt.Sprintf("%d/broadcast", nodeType)
}
