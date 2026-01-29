package service

import (
	"fmt"

	"github.com/hechh/framework"
	"github.com/hechh/framework/bus/internal/entity"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/yaml"

	"google.golang.org/protobuf/proto"
)

type Service struct {
	conn framework.IMessage
}

func NewService() *Service {
	return &Service{}
}

func (d *Service) Init(cfg *yaml.NatsConfig) error {
	conn, err := entity.NewNatsBus(cfg.Topic, cfg.Endpoints)
	if err != nil {
		return err
	}
	d.conn = conn
	return nil
}

func (d *Service) Close() {
	d.conn.Close()
}

func (d *Service) broadcastTopic(nodeType uint32) string {
	return fmt.Sprintf("%d", nodeType)
}

func (d *Service) sendTopic(nodeType uint32, nodeId uint32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

func (d *Service) replyTopic(nodeType uint32, nodeId uint32) string {
	return fmt.Sprintf("reply/%d/%d", nodeType, nodeId)
}

// 监听广播
func (d *Service) SubscribeBroadcast(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.broadcastTopic(framework.GetSelfType()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		err := proto.Unmarshal(msg.Body, pack)
		mlog.Trace(-1, "[Nats] 接收广播消息：head:%v, body:%d, error:%v", pack.Head, len(msg.Body), err)
		if err != nil {
			mlog.Error(0, "解析广播数据包错误:%v", err)
			return
		}
		f(pack.Head, pack.Body)
	})
}

// 监听单播
func (d *Service) SubscribeUnicast(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.sendTopic(framework.GetSelfType(), framework.GetSelfId()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		err := proto.Unmarshal(msg.Body, pack)
		mlog.Trace(-1, "[Nats] 接收单播消息：head:%v, body:%d, error:%v, router:%v", pack.Head, len(msg.Body), err, pack.List)
		if err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		// 更新路由
		for _, rr := range pack.List {
			framework.GetOrNewRouter(rr.GetIdType(), rr.GetId()).SetRouter(rr.List...)
		}
		mlog.Trace(-1, "[router] 玩家%d路由表%v", pack.Head.Id, pack.List)
		f(pack.Head, pack.Body)
	})
}

// 监听同步请求
func (d *Service) SubscribeReply(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.replyTopic(framework.GetSelfType(), framework.GetSelfId()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		err := proto.Unmarshal(msg.Body, pack)
		if pack.Head != nil {
			pack.Head.Reply = msg.Reply
		}
		mlog.Trace(-1, "[Nats] 接收同步消息：head:%v, body:%d, error:%v, router:%v", pack.Head, len(msg.Body), err, pack.List)
		if err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		// 更新路由
		for _, rr := range pack.List {
			framework.GetOrNewRouter(rr.GetIdType(), rr.GetId()).SetRouter(rr.List...)
		}
		mlog.Trace(-1, "[router] 玩家%d路由表%v", pack.Head.Id, pack.List)
		f(pack.Head, pack.Body)
	})
}

// 发送广播
func (d *Service) Broadcast(pack *packet.Packet) error {
	buf, err := proto.Marshal(pack)
	mlog.Trace(-1, "[Nats] 发送广播消息：head:%v, body:%d, error:%v", pack.Head, len(buf), err)
	if err != nil {
		return err
	}
	return d.conn.Send(d.broadcastTopic(pack.Head.DstNodeType), buf)
}

// 发送请求
func (d *Service) Send(pack *packet.Packet) error {
	buf, err := proto.Marshal(pack)
	mlog.Trace(-1, "[Nats] 发送单播消息：head:%v, body:%d, error:%v", pack.Head, len(buf), err)
	if err != nil {
		return err
	}
	return d.conn.Send(d.sendTopic(pack.Head.DstNodeType, pack.Head.DstNodeId), buf)
}

// 同步请求
func (d *Service) Request(pack *packet.Packet, cb func([]byte) error) error {
	buf, err := proto.Marshal(pack)
	mlog.Trace(-1, "[Nats] 发送同步消息：head:%v, body:%d, error:%v", pack.Head, len(buf), err)
	if err != nil {
		return err
	}
	return d.conn.Request(d.replyTopic(pack.Head.DstNodeType, pack.Head.DstNodeId), buf, cb)
}

// 同步应答
func (d *Service) Response(head *packet.Head, body []byte) error {
	err := d.conn.Response(head.Reply, body)
	mlog.Trace(-1, "[Nats] 发送同步回复：head:%v, body:%d, error:%v", head, len(body), err)
	return err
}
