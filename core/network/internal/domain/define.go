package domain

import (
	"fmt"
	"sync/atomic"

	"github.com/hechh/framework/packet"
)

const (
	IDLE_INTERVAL = 30 // 空闲超时（秒）
)

var (
	socketId   atomic.Uint32
	packetFunc func(*packet.Packet) error
	decodeFunc func([]byte) (*packet.Packet, error)
	encodeFunc func(*packet.Packet) ([]byte, error)
)

func GenSocketId() uint32 {
	return socketId.Add(1)
}

func SetDecodeFunc(f func([]byte) (*packet.Packet, error)) {
	decodeFunc = f
}

func SetEncodeFunc(f func(*packet.Packet) ([]byte, error)) {
	encodeFunc = f
}

func SetPacketFunc(f func(*packet.Packet) error) {
	packetFunc = f
}

func DecodeFrame(body []byte) (*packet.Packet, error) {
	if decodeFunc != nil {
		return decodeFunc(body)
	}
	return nil, fmt.Errorf("数据帧解码函数不存在")
}

func EncodeFrame(p *packet.Packet) ([]byte, error) {
	if encodeFunc != nil {
		return encodeFunc(p)
	}
	return nil, fmt.Errorf("数据帧编码函数不存在")
}

func PacketHandler(p *packet.Packet) error {
	if packetFunc != nil {
		return packetFunc(p)
	}
	return fmt.Errorf("数据包处理函数不存在")
}
