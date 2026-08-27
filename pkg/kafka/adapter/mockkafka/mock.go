package mockkafka

import (
	"sync"
	"time"

	"github.com/hechh/framework/pkg/kafka"
)

// Mock 内存版 Kafka 适配器（进程内发布/订阅），供单元测试使用，
// 无需真实 broker。实现 kafka.IKafka 接口。
type Mock struct {
	mu       sync.Mutex
	messages []*kafka.Message // 消息队列（先入先出）
	cfg      *kafka.Config
}

func NewMock() *Mock {
	return &Mock{}
}

func (d *Mock) Init(cfg *kafka.Config) error {
	d.cfg = cfg
	return nil
}

func (d *Mock) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.messages = nil
}

// Subscribe mock 无需真实订阅关系，记录配置即可。
func (d *Mock) Subscribe(topics []string) error {
	return nil
}

// Send 异步发送消息（入队）。
func (d *Mock) Send(topic string, body []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.messages = append(d.messages, &kafka.Message{Topic: topic, Value: body})
	return nil
}

// SendSync 同步发送消息（mock 中与 Send 等价）。
func (d *Mock) SendSync(topic string, body []byte) error {
	return d.Send(topic, body)
}

// ReadMessage 读取一条消息，队列为空且超时则返回 kafka.ErrPollTimeout。
func (d *Mock) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	deadline := time.Now().Add(timeout)
	for {
		d.mu.Lock()
		if len(d.messages) > 0 {
			msg := d.messages[0]
			d.messages = d.messages[1:]
			d.mu.Unlock()
			return msg, nil
		}
		d.mu.Unlock()

		if time.Now().After(deadline) {
			return nil, kafka.ErrPollTimeout
		}
		time.Sleep(5 * time.Millisecond)
	}
}
