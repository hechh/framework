package msgbus

import (
	"fmt"

	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/msgbus/internal/base"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/logic"
	"github.com/hechh/library/base/uerror"
	"github.com/hechh/library/mlog"
	"google.golang.org/protobuf/proto"
)

var object *MsgBus

func SetObject(o *MsgBus) {
	object = o
}

func SetPacketFunc(f func(*packet.Packet)) {
	base.SetPacketFunc(f)
}

// Subscribe 订阅主题
func Subscribe(topic string, f func(*packet.Packet)) error {
	if object != nil {
		return object.Subscribe(topic, f)
	}
	return fmt.Errorf("消息队列未初始化")
}

// Publish 发布消息到指定主题
func Publish(topic string, body []byte) error {
	if object != nil {
		return object.Publish(topic, body)
	}
	return fmt.Errorf("消息队列未初始化")
}

// 发送同步响应消息
func Response(reply string, msg proto.Message) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, msg)
	defer base.PutBytes(body)
	if err != nil {
		mlog.Errorf("MsgQueue.Response 参数序列化失败 error:%v, msg:%v", err, msg)
		return err
	}
	return object.Response(reply, body)
}

// 发送同步请求
func Request(head *packet.Head, msg proto.Message, rsp proto.Message, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, msg)
	defer base.PutBytes(body)
	if err != nil {
		mlog.Errorf("MsgQueue.Request 参数序列化失败 error:%v, msg:%v", err, msg)
		return err
	}
	return object.Request(head, body, rsp, funcs...)
}

func RequestRaw(head *packet.Head, msg []byte, rsp proto.Message, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	return object.Request(head, msg, rsp, funcs...)
}

// 发送广播消息
func Broadcast(head *packet.Head, msg proto.Message, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, msg)
	defer base.PutBytes(body)
	if err != nil {
		mlog.Errorf("MsgQueue.Broadcast 参数序列化失败 error:%v, msg:%v", err, msg)
		return err
	}
	return object.Broadcast(head, body, funcs...)
}

func BroadcastRaw(head *packet.Head, msg []byte, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	return object.Broadcast(head, msg, funcs...)
}

// 发送单播消息
func Send(head *packet.Head, msg proto.Message, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, msg)
	defer base.PutBytes(body)
	if err != nil {
		mlog.Errorf("MsgQueue.Send 参数序列化失败 error:%v, msg:%v", err, msg)
		return err
	}
	return object.Send(head, body, funcs...)
}

func SendRaw(head *packet.Head, msg []byte, funcs ...func(*packet.Packet) error) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	return object.Send(head, msg, funcs...)
}

func SendToClient(head *packet.Head, msg proto.Message) error {
	return Send(head, msg, fun.SetRspClient, fun.CacheRouting)
}

func NotifyToClient(head *packet.Head, msg proto.Message, uids ...uint64) error {
	if object == nil {
		return fmt.Errorf("MsgQueue未初始化")
	}
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, msg)
	defer base.PutBytes(body)
	if err != nil {
		mlog.Errorf("MsgQueue.Send 参数序列化失败 error:%v, msg:%v", err, msg)
		return err
	}
	return object.Notify(head, body, uids)
}

func AutoRsp(ctx define.IContext, h handler.IHandler, head *packet.Head, rsp any, reterr error) {
	uerror.SetRspHead(rsp, reterr)

	irsp, ok := rsp.(proto.Message)
	if !ok {
		mlog.Errorf("跨服务转发只支持protobuf协议 func=%s", h.GetActorFuncName())
		return
	}

	var err error
	if head.Reply != "" {
		err = Response(head.Reply, irsp)
	} else if head.Cmd > 0 {
		err = Send(head, irsp, fun.SetRspClient, fun.CacheRouting)
	} else if head.Back != nil {
		err = Send(head, irsp, fun.SetRspBack)
	}

	if err != nil {
		ctx.Error("自动回复失败", err, head, rsp)
	} else if !logic.Has(h.GetMask(), define.LOG_MASK) {
		ctx.Trace("自动回复成功", err, head, rsp)
	}
}
