package service

import (
	"fmt"
	"framework/define"
	"framework/internal/global"
	"framework/library/mlog"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/bus/internal/entity"
	"framework/repository/router"

	"github.com/golang/protobuf/proto"
)

type Service struct {
	conn define.IMsgQueue
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
	return d.conn.Subscribe(d.broadcastTopic(global.GetSelfType()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析广播数据包错误:%v", err)
			return
		}
		f(pack.Head, pack.Body)
	})
}

// 监听单播
func (d *Service) SubscribeUnicast(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.sendTopic(global.GetSelfType(), global.GetSelfId()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		for _, rr := range pack.List {
			router.GetOrNew(rr.GetIdType(), rr.GetId()).SetRouter(rr.List...)
		}
		f(pack.Head, pack.Body)
	})
}

// 监听同步请求
func (d *Service) SubscribeReply(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.replyTopic(global.GetSelfType(), global.GetSelfId()), func(msg *packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		for _, rr := range pack.List {
			router.GetOrNew(rr.GetIdType(), rr.GetId()).SetRouter(rr.List...)
		}
		f(pack.Head, pack.Body)
	})
}

// 发送广播
func (d *Service) Broadcast(pack *packet.Packet) error {
	pack.Head.SrcNodeType = global.GetSelfType()
	pack.Head.SrcNodeId = global.GetSelfId()
	if pack.Head.Back != nil {
		pack.Head.Back.NodeType = global.GetSelfType()
		pack.Head.Back.NodeId = global.GetSelfId()
	}
	buf, err := proto.Marshal(pack)
	if err != nil {
		return err
	}
	return d.conn.Send(d.broadcastTopic(pack.Head.DstNodeType), buf)
}

// 发送请求
func (d *Service) Send(pack *packet.Packet) error {
	pack.Head.SrcNodeType = global.GetSelfType()
	pack.Head.SrcNodeId = global.GetSelfId()
	if pack.Head.Back != nil {
		pack.Head.Back.NodeType = global.GetSelfType()
		pack.Head.Back.NodeId = global.GetSelfId()
	}
	buf, err := proto.Marshal(pack)
	if err != nil {
		return err
	}
	return d.conn.Send(d.sendTopic(pack.Head.DstNodeType, pack.Head.DstNodeId), buf)
}

// 同步请求
func (d *Service) Request(pack *packet.Packet, cb func([]byte) error) error {
	pack.Head.SrcNodeType = global.GetSelfType()
	pack.Head.SrcNodeId = global.GetSelfId()
	if pack.Head.Back != nil {
		pack.Head.Back.NodeType = global.GetSelfType()
		pack.Head.Back.NodeId = global.GetSelfId()
	}
	buf, err := proto.Marshal(pack)
	if err != nil {
		return err
	}
	return d.conn.Request(d.replyTopic(pack.Head.DstNodeType, pack.Head.DstNodeId), buf, cb)
}

// 同步应答
func (d *Service) Response(head *packet.Head, body []byte) error {
	return d.conn.Response(head.Reply, body)
}
