package network

import (
	"fmt"

	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/core/network/internal/frame"
	"github.com/hechh/framework/packet"
)

var object *Network

func init() {
	domain.SetDecodeFunc(frame.Decode)
	domain.SetEncodeFunc(frame.Encode)
}

// SetObject 注入全局 Network 实例
func SetObject(obj *Network) {
	object = obj
}

func SetDecodeFunc(f func([]byte) (*packet.Packet, error)) {
	domain.SetDecodeFunc(f)
}

func SetEncodeFunc(f func(*packet.Packet) ([]byte, error)) {
	domain.SetEncodeFunc(f)
}

func SetPacketFunc(f func(*packet.Packet) error) {
	domain.SetPacketFunc(f)
}

// Bind 绑定 socketId ↔ uid（全局便捷方法）
func Bind(socketId uint32, uid uint64) bool {
	if object != nil {
		return object.Bind(socketId, uid)
	}
	return false
}

// Unbind 解绑并移除连接（全局便捷方法）
func Unbind(socketId uint32, uid uint64) {
	if object != nil {
		object.Unbind(socketId, uid)
	}
}

// SendToClient 发送消息到客户端（全局便捷方法）
func Send(head *packet.Head, body []byte) error {
	if object != nil {
		return object.Send(head, body)
	}
	return fmt.Errorf("Network未初始化")
}
