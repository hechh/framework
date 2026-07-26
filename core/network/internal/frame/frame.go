package frame

import (
	"encoding/binary"
	"fmt"

	"github.com/hechh/framework/packet"
)

// 帧格式常量
const (
	FrameHeadLen        = 4       // 包体长度字段(4字节)
	PacketHeadLen       = 20      // 包头：cmd(4) + version(4) + seq(4) + uid(8)
	DefaultMaxFrameSize = 4194304 // 默认最大帧大小 4MB
)

// FrameCodec 实现 IFrameCodec 接口，负责数据帧的编解码
// 数据格式：[4字节body长度][4字节cmd][4字节version][4字节seq][8字节uid][body数据]

// Decode 解码数据帧，将字节流转为 packet.Packet
func Decode(data []byte) (*packet.Packet, error) {
	minLen := FrameHeadLen + PacketHeadLen
	if len(data) < minLen {
		return nil, fmt.Errorf("数据长度不足: %d < %d", len(data), minLen)
	}

	bodyLen := binary.BigEndian.Uint32(data[:FrameHeadLen])
	if bodyLen > DefaultMaxFrameSize {
		return nil, fmt.Errorf("包体长度超出限制: %d > %d", bodyLen, DefaultMaxFrameSize)
	}

	totalLen := FrameHeadLen + PacketHeadLen + int(bodyLen)
	if len(data) < totalLen {
		return nil, fmt.Errorf("数据不完整: 期望 %d, 实际 %d", totalLen, len(data))
	}

	offset := FrameHeadLen
	head := &packet.Head{
		Cmd:     binary.BigEndian.Uint32(data[offset : offset+4]),
		Version: binary.BigEndian.Uint32(data[offset+4 : offset+8]),
		Seq:     binary.BigEndian.Uint32(data[offset+8 : offset+12]),
		Uid:     binary.BigEndian.Uint64(data[offset+12 : offset+20]),
	}

	var body []byte
	if bodyLen > 0 {
		body = data[FrameHeadLen+PacketHeadLen : totalLen]
	}

	return &packet.Packet{
		Head: head,
		Body: body,
	}, nil
}

// Encode 编码数据帧，将 packet.Packet 转为字节流
func Encode(pack *packet.Packet) ([]byte, error) {
	if pack == nil {
		return nil, fmt.Errorf("pack is nil")
	}
	if pack.Head == nil {
		return nil, fmt.Errorf("pack.Head is nil")
	}

	bodyLen := uint32(len(pack.Body))
	if bodyLen > DefaultMaxFrameSize {
		return nil, fmt.Errorf("包体长度超出限制: %d > %d", bodyLen, DefaultMaxFrameSize)
	}

	totalLen := FrameHeadLen + PacketHeadLen + int(bodyLen)
	data := make([]byte, totalLen)

	binary.BigEndian.PutUint32(data[0:4], bodyLen)

	offset := FrameHeadLen
	binary.BigEndian.PutUint32(data[offset:offset+4], pack.Head.Cmd)
	binary.BigEndian.PutUint32(data[offset+4:offset+8], pack.Head.Version)
	binary.BigEndian.PutUint32(data[offset+8:offset+12], pack.Head.Seq)
	binary.BigEndian.PutUint64(data[offset+12:offset+20], pack.Head.Uid)
	if bodyLen > 0 {
		copy(data[FrameHeadLen+PacketHeadLen:], pack.Body)
	}
	return data, nil
}
