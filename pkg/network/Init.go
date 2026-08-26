package network

import (
	"fmt"

	"github.com/hechh/framework/library/enum"
	"github.com/hechh/framework/library/uerror"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/network/internal/domain"
	"github.com/hechh/framework/pkg/network/internal/frame"
)

var (
	object       *Network
	cmdConvertor enum.IConvertor
)

func init() {
	domain.SetDecodeFunc(frame.Decode)
	domain.SetEncodeFunc(frame.Encode)
}

// SetObject 注入全局 Network 实例
func SetObject(obj *Network) {
	object = obj
}

func SetCmdConvertor(n map[string]int32, i map[int32]string) {
	cmdConvertor = enum.WrapConvertor(n, i)
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

func SendToClient(head *packet.Head, err error, rsp IMessage) error {
	uerror.SetRspHead(rsp, err)
	body, err := rsp.MarshalVT()
	if err != nil {
		mlog.Errorf("SendRspToClient: MarshalVT失败, err=%v", err)
		return err
	}
	return SendRawToClient(head, body)
}

// SendToClient 发送消息到客户端（全局便捷方法）
func SendRawToClient(head *packet.Head, body []byte) error {
	if object == nil {
		return fmt.Errorf("Network未初始化")
	}
	if head.Cmd%2 == 0 {
		if cmdConvertor.Has(head.Cmd + 1) {
			head.Cmd++
		}
	}
	return object.Send(head, body)
	/*
		if err != nil {
			mlog.Tracef("SendRawToClient: 向客户端发送响应, cmd=%d, uid=%d, body=%d, error=%v", head.Cmd, head.Uid, len(body), err)
		} else {
			mlog.Tracef("SendRawToClient: 向客户端发送响应, cmd=%d, uid=%d, body=%d", head.Cmd, head.Uid, len(body))
		}
		return err
	*/
}
