package service

import (
	"fmt"
	"framework/define"
	"framework/internal/global"
	"framework/internal/router"
	"framework/library/mlog"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/bus/internal/entity"

	"github.com/gogo/protobuf/proto"
)

type BusService struct {
	conn define.IBus
}

func NewBusService(cfg *yaml.NatsConfig) (*BusService, error) {
	conn, err := entity.NewNatsBus(cfg.Topic, cfg.Endpoints)
	if err != nil {
		return nil, err
	}
	return &BusService{conn: conn}, nil
}

func (d *BusService) Close() {
	d.conn.Close()
}

func (d *BusService) broadcastTopic(nodeType int32) string {
	return fmt.Sprintf("%d", nodeType)
}

func (d *BusService) readTopic(nodeType int32, nodeId int32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

func (d *BusService) replyTopic(nodeType int32, nodeId int32) string {
	return fmt.Sprintf("reply/%d/%d", nodeType, nodeId)
}

// 监听广播
func (d *BusService) Listen(f func(head *packet.Head, body []byte)) error {
	return d.conn.Read(d.broadcastTopic(global.GetSelfType()), func(msg packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析广播数据包错误:%v", err)
			return
		}
		f(pack.Head, pack.Body)
	})
}

// 发送广播
func (d *BusService) Broadcast(head *packet.Head, body []byte, rs ...*packet.Router) error {
	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Write(d.broadcastTopic(head.DstNodeType), buf)
}

// 监听单播
func (d *BusService) Read(f func(head *packet.Head, body []byte)) error {
	return d.conn.Read(d.readTopic(global.GetSelfType(), global.GetSelfId()), func(msg packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		router.SetRouter(pack.List...)
		f(pack.Head, pack.Body)
	})
}

// 发送请求
func (d *BusService) Write(head *packet.Head, body []byte, rs ...*packet.Router) error {
	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Write(d.readTopic(head.DstNodeType, head.DstNodeId), buf)
}

// 监听同步请求
func (d *BusService) Reply(f func(head *packet.Head, body []byte)) error {
	return d.conn.Read(d.replyTopic(global.GetSelfType(), global.GetSelfId()), func(msg packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		router.SetRouter(pack.List...)
		f(pack.Head, pack.Body)
	})
}

// 同步请求
func (d *BusService) Request(cb func([]byte) error, head *packet.Head, body []byte, rs ...*packet.Router) error {
	buf, err := proto.Marshal(&packet.Packet{Head: head, Body: body, List: rs})
	if err != nil {
		return err
	}
	return d.conn.Request(d.replyTopic(head.DstNodeType, head.DstNodeId), buf, cb)
}

// 同步应答
func (d *BusService) Response(head *packet.Head, body []byte) error {
	return d.conn.Response(head.Reply, body)
}
