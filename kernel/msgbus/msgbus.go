package msgbus

import (
	"fmt"

	"github.com/hechh/framework/kernel/global"
	"github.com/hechh/framework/kernel/msgbus/internal/base"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
	"google.golang.org/protobuf/proto"
)

type NatsConfig struct {
	Prefix        string `yaml:"prefix,omitempty"`
	Endpoints     string `yaml:"endpoints,omitempty"`
	MaxReconnect  int32  `yaml:"max_reconnect,omitempty"`
	ReconnectWait int64  `yaml:"reconnect_wait,omitempty"`
	PingInterval  int64  `yaml:"ping_interval,omitempty"`
	DrainTimeout  int64  `yaml:"drain_timeout,omitempty"`
}

type Config struct {
	Nats *NatsConfig `yaml:"nats,omitempty"`
}

type IMessage interface {
	Init(*Config) error                                             // 初始化
	Close()                                                         // 关闭消息队列
	Subscribe(topic string, handle func(*packet.Message)) error     // 读取消息
	Publish(topic string, body []byte) error                        // 发布消息到指定主题
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
}

type MsgBus struct {
	adapter       IMessage
	cfg           *Config
	selfPoint     string // 当前节点单播主题
	selfBroadcast string // 当前节点广播主题
	selfReply     string // 当前节点回复主题
}

func NewMsgBus(msg IMessage) *MsgBus {
	return &MsgBus{
		adapter: msg,
	}
}

func (d *MsgBus) Init(cfg *Config) error {
	if err := d.adapter.Init(cfg); err != nil {
		return err
	}

	// 预计算当前节点主题
	d.cfg = cfg
	nodeType := global.GetSelfNodeType()
	nodeId := global.GetSelfNodeId()
	d.selfPoint = base.BuildPoint(nodeType, nodeId)
	d.selfBroadcast = base.BuildBroadcast(nodeType)
	d.selfReply = base.BuildReply(nodeType, nodeId)

	fun := func(msg *packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Errorf("[nats] 反序列化消息失败: %v", err)
			return
		}
		pack.Head.Reply = msg.Reply
		base.PacketHandler(pack)
	}

	// 订阅广播主题
	if err := d.adapter.Subscribe(d.selfBroadcast, fun); err != nil {
		return err
	}
	// 订阅单播主题
	if err := d.adapter.Subscribe(d.selfPoint, fun); err != nil {
		return err
	}
	// 订阅回复主题
	if err := d.adapter.Subscribe(d.selfReply, fun); err != nil {
		return err
	}
	return nil
}

func (d *MsgBus) Close() {
	d.adapter.Close()
}

// Subscribe 订阅主题
func (d *MsgBus) Subscribe(topic string, f func(*packet.Packet)) error {
	return d.adapter.Subscribe(topic, func(msg *packet.Message) {
		pack := &packet.Packet{}
		if err := proto.Unmarshal(msg.Body, pack); err != nil {
			mlog.Errorf("[nats] 反序列化消息失败: %v", err)
			return
		}
		pack.Head.Reply = msg.Reply
		f(pack)
	})
}

// Publish 发布消息到指定主题
func (d *MsgBus) Publish(topic string, body []byte) error {
	return d.adapter.Publish(topic, body)
}

// Broadcast 广播消息
func (d *MsgBus) Broadcast(head *packet.Head, msg []byte, funcs ...func(*packet.Packet) error) error {
	head.SrcType = global.GetSelfNodeType()
	head.SrcId = global.GetSelfNodeId()
	pack := &packet.Packet{
		Head: head,
		Body: msg,
	}

	// 设置packet
	for i, f := range funcs {
		if err := f(pack); err != nil {
			mlog.Errorf("MsgQueue.Broadcast参数错误 position:%d, error:%v", i, err)
			return err
		}
	}

	// 序列化
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, pack)
	defer base.PutBytes(body)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// 发送
	return d.adapter.Publish(base.BuildBroadcast(pack.Head.DstType), body)
}

// Response 响应同步消息
func (d *MsgBus) Response(reply string, body []byte) error {
	return d.adapter.Response(reply, body)
}

// Request 发送同步请求
func (d *MsgBus) Request(head *packet.Head, msg []byte, rsp proto.Message, funcs ...func(*packet.Packet) error) error {
	head.SrcType = global.GetSelfNodeType()
	head.SrcId = global.GetSelfNodeId()
	pack := &packet.Packet{
		Head: head,
		Body: msg,
	}

	// 设置packet
	for i, f := range funcs {
		if err := f(pack); err != nil {
			mlog.Errorf("MsgQueue.Broadcast参数错误 position:%d, error:%v", i, err)
			return err
		}
	}

	// 序列化
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, pack)
	defer base.PutBytes(body)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// 发送消息
	topic := base.BuildReply(pack.Head.DstType, pack.Head.DstId)
	return d.adapter.Request(topic, body, func(respBody []byte) error {
		return proto.Unmarshal(respBody, rsp)
	})
}

// Send 发送点对点消息
func (d *MsgBus) Send(head *packet.Head, msg []byte, funcs ...func(*packet.Packet) error) error {
	head.SrcType = global.GetSelfNodeType()
	head.SrcId = global.GetSelfNodeId()
	pack := &packet.Packet{
		Head: head,
		Body: msg,
	}

	// 设置packet
	for i, f := range funcs {
		if err := f(pack); err != nil {
			mlog.Errorf("MsgQueue.Broadcast参数错误 position:%d, error:%v", i, err)
			return err
		}
	}

	// 序列化
	body := base.GetBytes()
	body, err := proto.MarshalOptions{}.MarshalAppend(body, pack)
	defer base.PutBytes(body)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}

	// 发送消息
	return d.adapter.Publish(base.BuildPoint(pack.Head.DstType, pack.Head.DstId), body)
}

// 发送通知
func (d *MsgBus) Notify(head *packet.Head, msg []byte, uids []uint64, funcs ...func(*packet.Packet) error) error {
	if head.Uid > 0 {
		uids = append(uids, head.Uid)
	}

	tmps := map[uint64]struct{}{}
	for _, uid := range uids {
		if _, ok := tmps[uid]; ok {
			continue
		}
		tmps[uid] = struct{}{}
		head.Uid = uid

		// 重置
		if err := d.Send(head, msg, funcs...); err != nil {
			mlog.Errorf("MsgBus.Notify 发送单播消息失败 uid=%d, error=%v", uid, err)
		}
	}
	return nil
}
