package service

import (
	"fmt"
	"framework/core/bus/domain"
	"framework/core/bus/internal/entity"
	"framework/core/router"
	"framework/library/mlog"
	"framework/library/yaml"
	"framework/packet"

	"github.com/gogo/protobuf/proto"
)

type Service struct {
	self *packet.Node
	conn domain.IMsgQueue
}

func NewService() *Service {
	return &Service{}
}

func (d *Service) Init(cfg *yaml.NatsConfig, nn *packet.Node) error {
	conn, err := entity.NewNatsBus(cfg.Topic, cfg.Endpoints)
	if err != nil {
		return err
	}
	d.conn = conn
	d.self = nn
	return nil
}

func (d *Service) Close() {
	d.conn.Close()
}

func (d *Service) broadcastTopic(nodeType uint32) string {
	return fmt.Sprintf("%d", nodeType)
}

func (d *Service) readTopic(nodeType uint32, nodeId uint32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

func (d *Service) replyTopic(nodeType uint32, nodeId uint32) string {
	return fmt.Sprintf("reply/%d/%d", nodeType, nodeId)
}

// 监听广播
func (d *Service) SubscribeBroadcast(f func(head *packet.Head, body []byte)) error {
	return d.conn.Subscribe(d.broadcastTopic(d.self.Type), func(msg packet.Message) {
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
	return d.conn.Subscribe(d.readTopic(d.self.Type, d.self.Id), func(msg packet.Message) {
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
	return d.conn.Subscribe(d.replyTopic(d.self.Type, d.self.Id), func(msg packet.Message) {
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
func (d *Service) Broadcast(head *packet.Head, body []byte, rs ...*packet.Router) error {
	head.SrcNodeType = d.self.Type
	head.SrcNodeId = d.self.Id
	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Send(d.broadcastTopic(head.DstNodeType), buf)
}

// 发送请求
func (d *Service) Send(head *packet.Head, body []byte, rs ...*packet.Router) error {
	head.SrcNodeType = d.self.Type
	head.SrcNodeId = d.self.Id
	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Send(d.readTopic(head.DstNodeType, head.DstNodeId), buf)
}

// 同步请求
func (d *Service) Request(cb func([]byte) error, head *packet.Head, body []byte, rs ...*packet.Router) error {
	head.SrcNodeType = d.self.Type
	head.SrcNodeId = d.self.Id

	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Request(d.replyTopic(head.DstNodeType, head.DstNodeId), buf, cb)
}

// 同步应答
func (d *Service) Response(head *packet.Head, body []byte) error {
	return d.conn.Response(head.Reply, body)
}
