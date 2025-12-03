package entity

import (
	"fmt"
	"framework/define"
	"framework/internal/global"
	"framework/library/mlog"
	"framework/packet"

	"github.com/golang/protobuf/proto"
)

type Bus struct {
	conn define.IBus
}

func NewBus(cc define.IBus) *Bus {
	return &Bus{conn: cc}
}

func (d *Bus) Close() {
	d.conn.Close()
}

func (d *Bus) topicBroadcast(nodeType int32) string {
	return fmt.Sprintf("%d", nodeType)
}

func (d *Bus) topicPoint(nodeType int32, nodeId int32) string {
	return fmt.Sprintf("%d/%d", nodeType, nodeId)
}

func (d *Bus) topicReply(nodeType int32, nodeId int32) string {
	return fmt.Sprintf("reply/%d/%d", nodeType, nodeId)
}

func (d *Bus) ReadBroadcast(f func(packet.Packet)) error {
	return d.conn.Read(d.topicBroadcast(global.GetSelfType()), func(msg packet.Message) {
		pack := packet.Packet{}
		if err := proto.Unmarshal(msg.Body, &pack); err != nil {
			mlog.Error(0, "解析广播数据包错误:%v", err)
			return
		}
		f(pack)
	})
}

func (d *Bus) Broadcast(pac packet.Packet) error {
	buf, err := proto.Marshal(&pac)
	if err != nil {
		return err
	}
	return d.conn.Write(d.topicBroadcast(pac.Head.DstNodeType), buf)
}

func (d *Bus) Read(f func(packet.Packet)) error {
	return d.conn.Read(d.topicPoint(global.GetSelfType(), global.GetSelfId()), func(msg packet.Message) {
		pack := packet.Packet{}
		if err := proto.Unmarshal(msg.Body, &pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		f(pack)
	})
}

func (d *Bus) Write(pac *packet.Packet) error {
	buf, err := proto.Marshal(pac)
	if err != nil {
		return err
	}
	return d.conn.Write(d.topicPoint(pac.Head.DstNodeType, pac.Head.DstNodeId), buf)
}

func (d *Bus) ReadReply(f func(packet.Packet)) error {
	return d.conn.Read(d.topicReply(global.GetSelfType(), global.GetSelfId()), func(msg packet.Message) {
		pack := packet.Packet{}
		if err := proto.Unmarshal(msg.Body, &pack); err != nil {
			mlog.Error(0, "解析单播数据包错误:%v", err)
			return
		}
		pack.Head.Reply = msg.Reply
		f(pack)
	})
}

func (d *Bus) Request(pac *packet.Packet, cb func([]byte) error) error {
	buf, err := proto.Marshal(pac)
	if err != nil {
		return err
	}
	return d.conn.Request(d.topicReply(pac.Head.DstNodeType, pac.Head.DstNodeId), buf, cb)
}

func (d *Bus) Response(topic string, body []byte) error {
	return d.conn.Response(topic, body)
}
