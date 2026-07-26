package domain

import (
	"fmt"

	"github.com/hechh/framework/packet"
)

const (
	IDLE_INTERVAL = 30 // 空闲超时（秒）
)

/*
// 数据帧编码/解码
type IFrame interface {
	Decode([]byte) (*packet.Packet, error)
	Encode(*packet.Packet) ([]byte, error)
}
*/

// IClient 网络客户端接口
type IClient interface {
	Init()                           // 启动客户端读写循环
	Close()                          // 关闭客户端
	GetId() uint32                   // 获取 socketId
	GetUid() uint64                  // 获取uid
	SetUid(uint64) bool              // 设置uid
	GetUpdateTime() int64            // 获取最后活跃时间（Unix 秒）
	Send(*packet.Head, []byte) error // 发送消息
}

// IServer 网络服务端接口
type IServer interface {
	Init(*packet.Config) error
	Close()
	Bind(uint32, uint64) bool
	Add(IClient)
	Get(any) IClient
	Del(any)
}

var (
	packetFunc func(*packet.Packet) error
	decodeFunc func([]byte) (*packet.Packet, error)
	encodeFunc func(*packet.Packet) ([]byte, error)
)

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
