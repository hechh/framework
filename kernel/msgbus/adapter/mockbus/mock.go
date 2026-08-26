package mockbus

import (
	"fmt"
	"time"

	"github.com/hechh/framework/kernel/msgbus"
	"github.com/hechh/framework/packet"
	natsrv "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// MockMessage 内嵌 NATS 服务器的消息队列实现，用于测试。
// 启动一个进程内真实的 NATS Server，所有 Publish/Subscribe/Request/Response 均走真实链路。
type MockMessage struct {
	ns     *natsrv.Server       // 内嵌 NATS 服务器进程
	client *nats.Conn           // 连接到内嵌服务器的客户端
	subs   []*nats.Subscription // 所有订阅，用于 Close 时清理
}

// NewMockMessage 创建一个新的 MockMessage 实例。
func New() *MockMessage {
	return &MockMessage{}
}

// Init 启动内嵌 NATS 服务器并建立客户端连接。
// cfg 参数保留以兼容接口，当前实现忽略配置，始终在 127.0.0.1 随机端口启动。
func (m *MockMessage) Init(cfg *msgbus.Config) error {
	// 创建内嵌 NATS 服务器，端口设为 -1 由 OS 自动分配
	opts := &natsrv.Options{
		Host: "127.0.0.1",
		Port: -1,
	}
	ns, err := natsrv.NewServer(opts)
	if err != nil {
		return fmt.Errorf("[mock] 创建 NATS 服务器失败: %w", err)
	}
	ns.Start()

	// 等待服务器就绪（最多 5 秒）
	if !ns.ReadyForConnections(5 * time.Second) {
		ns.Shutdown()
		return fmt.Errorf("[mock] NATS 服务器启动超时")
	}
	m.ns = ns

	// 连接到内嵌服务器
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		ns.Shutdown()
		return fmt.Errorf("[mock] 连接 NATS 服务器失败: %w", err)
	}
	m.client = nc
	return nil
}

// Close 优雅关闭：先取消所有订阅，再关闭客户端连接，最后关闭内嵌服务器。
func (m *MockMessage) Close() {
	// 取消所有订阅
	for _, sub := range m.subs {
		sub.Unsubscribe()
	}
	m.subs = nil

	// 关闭客户端连接
	if m.client != nil {
		m.client.Close()
		m.client = nil
	}

	// 关闭内嵌服务器
	if m.ns != nil {
		m.ns.Shutdown()
		m.ns = nil
	}
}

// Subscribe 订阅指定主题的消息。
func (m *MockMessage) Subscribe(topic string, handle func(*packet.Message)) error {
	sub, err := m.client.Subscribe(topic, func(msg *nats.Msg) {
		handle(&packet.Message{Body: msg.Data, Reply: msg.Reply})
	})
	if err != nil {
		return fmt.Errorf("[mock] 订阅失败 topic=%s: %w", topic, err)
	}
	m.subs = append(m.subs, sub)
	return nil
}

// Publish 发布消息到指定主题。
func (m *MockMessage) Publish(topic string, body []byte) error {
	return m.client.Publish(topic, body)
}

// Request 发送同步请求并回调处理响应。
func (m *MockMessage) Request(topic string, body []byte, cb func([]byte) error) error {
	resp, err := m.client.Request(topic, body, 5*time.Second)
	if err != nil {
		return fmt.Errorf("[mock] 请求失败 topic=%s: %w", topic, err)
	}
	return cb(resp.Data)
}

// Response 回复同步消息。
func (m *MockMessage) Response(topic string, body []byte) error {
	return m.client.Publish(topic, body)
}
