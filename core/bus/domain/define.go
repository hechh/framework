package domain

import "framework/packet"

// 消息队列接口
type IMsgQueue interface {
	Subscribe(topic string, handle func(packet.Message)) error      // 读取消息
	Send(topic string, body []byte) error                           // 发送消息
	Request(topic string, body []byte, cb func([]byte) error) error // 发送同步消息
	Response(topic string, body []byte) error                       // 回复同步消息
	Close()                                                         // 关闭消息总线服务
}

// 消息总线
type IBus interface {
	Listen(func(*packet.Head, []byte)) error                                   // 监听广播
	Broadcast(*packet.Head, []byte, ...*packet.Router) error                   // 发送广播
	Read(func(*packet.Head, []byte)) error                                     // 接受请求
	Write(*packet.Head, []byte, ...*packet.Router) error                       // 发送请求
	Reply(func(*packet.Head, []byte)) error                                    // 接受同步请求
	Request(func([]byte) error, *packet.Head, []byte, ...*packet.Router) error // 发送同步请求
	Response(*packet.Head, []byte) error                                       // 应答
}
