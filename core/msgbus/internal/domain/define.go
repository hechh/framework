package domain

import (
	"github.com/hechh/framework/packet"
)

const (
	DEFAULT_REQUEST_TIMEOUT = 5 // 默认同步请求超时（秒）
)

var (
	packetFunc func(*packet.Packet)
)

func SetPacketFunc(f func(*packet.Packet)) {
	packetFunc = f
}

func PacketHandler(msg *packet.Packet) {
	if packetFunc != nil {
		packetFunc(msg)
	}
}

type IMessage interface {
	Init(*packet.MsgQueueConfig) error                              // 初始化
	Close()                                                         // 关闭消息队列
	Subscribe(topic string, handle func(*packet.Message)) error     // 读取消息
	Publish(topic string, body []byte) error                        // 发布消息到指定主题
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
}
