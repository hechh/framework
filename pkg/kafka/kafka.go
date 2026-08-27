package kafka

import (
	"errors"
	"sync"
	"time"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/mlog"
)

// ErrPollTimeout 消费轮询超时（无消息到达），消费循环据此继续轮询。
var ErrPollTimeout = errors.New("kafka 轮询超时")

// Config Kafka 配置（对应 yaml 中 kafka 节点）。
type Config struct {
	Brokers          string   `yaml:"brokers,omitempty"`            // broker 地址（逗号分隔）
	GroupId          string   `yaml:"group_id,omitempty"`           // 消费组 ID（为空时使用默认值）
	Topics           []string `yaml:"topics,omitempty"`             // 启动时订阅的主题列表
	AutoOffsetReset  string   `yaml:"auto_offset_reset,omitempty"`  // 无已提交 offset 时的起点：earliest/latest
	EnableAutoCommit bool     `yaml:"enable_auto_commit,omitempty"` // 是否自动提交消费 offset
	SessionTimeoutMs int64    `yaml:"session_timeout_ms,omitempty"` // 会话超时（毫秒）
	PollTimeoutMs    int64    `yaml:"poll_timeout_ms,omitempty"`    // 消费轮询超时（毫秒），默认 100
	SendTimeoutMs    int64    `yaml:"send_timeout_ms,omitempty"`    // 同步发送等待确认超时（毫秒），默认 5000
}

// Message Kafka 消息（适配器统一返回给上层的类型）。
type Message struct {
	Topic  string // 主题
	Key    []byte // 消息键（可选）
	Value  []byte // 消息体
	Offset int64  // 分区偏移量（消费时有效）
}

// IHandler 主题消息处理函数。返回 error 仅记录日志，不中断消费循环。
type IHandler func(topic string, body []byte) error

// IKafka 底层消息队列适配器接口。
type IKafka interface {
	Init(*Config) error
	Close()
	Subscribe(topics []string) error                     // 订阅主题
	Send(topic string, body []byte) error                // 异步发送消息
	SendSync(topic string, body []byte) error            // 同步发送（等待投递确认）
	ReadMessage(timeout time.Duration) (*Message, error) // 读取一条消息（消费循环使用）
}

// Kafka 主题消息队列封装：订阅主题、循环消费并分发到处理器、发送消息。
type Kafka struct {
	adapter  IKafka
	cfg      *Config
	mu       sync.RWMutex
	handlers map[string]IHandler // 主题 → 消息处理器
	exitCh   chan struct{}
	wg       sync.WaitGroup
	started  bool
}

// NewKafka 创建 Kafka 封装（传入具体适配器，如 kafkago.NewKafkaGo()）。
func NewKafka(adapter IKafka) *Kafka {
	return &Kafka{
		adapter:  adapter,
		handlers: make(map[string]IHandler),
		exitCh:   make(chan struct{}),
	}
}

// Init 初始化适配器并订阅配置中声明的主题。
func (d *Kafka) Init(cfg *Config) error {
	if cfg == nil {
		return errors.New("kafka 配置为空")
	}
	if cfg.Brokers == "" {
		return errors.New("kafka brokers 未配置")
	}
	d.cfg = cfg
	if err := d.adapter.Init(cfg); err != nil {
		return err
	}
	// 订阅配置中声明的主题（处理器可随后通过 Subscribe 注册）
	if len(cfg.Topics) > 0 {
		if err := d.adapter.Subscribe(cfg.Topics); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe 订阅主题并注册消息处理器。重复订阅同一主题会覆盖其处理器。
func (d *Kafka) Subscribe(topic string, handler IHandler) error {
	if handler == nil {
		return errors.New("kafka 消息处理器为空")
	}
	d.mu.Lock()
	d.handlers[topic] = handler
	d.mu.Unlock()
	return d.adapter.Subscribe([]string{topic})
}

// Send 异步发送消息到指定主题。
func (d *Kafka) Send(topic string, body []byte) error {
	return d.adapter.Send(topic, body)
}

// SendSync 同步发送消息（等待 broker 确认或超时）。
func (d *Kafka) SendSync(topic string, body []byte) error {
	return d.adapter.SendSync(topic, body)
}

// Start 启动消费循环（内部 goroutine：循环读取消息并分发到对应主题处理器）。
// 可在 Subscribe 注册处理器之后调用；重复调用幂等。
func (d *Kafka) Start() {
	d.mu.Lock()
	if d.started {
		d.mu.Unlock()
		return
	}
	d.started = true
	d.mu.Unlock()

	d.wg.Add(1)
	safe.SafeGo(mlog.Fatalf, func() {
		defer d.wg.Done()
		d.consumeLoop()
	})
	mlog.Infof("[kafka] 消费循环已启动")
}

// consumeLoop 消费循环：轮询读取消息，分发到对应主题处理器。
func (d *Kafka) consumeLoop() {
	poll := d.pollTimeout()
	for {
		select {
		case <-d.exitCh:
			return
		default:
		}

		msg, err := d.adapter.ReadMessage(poll)
		if err != nil {
			if errors.Is(err, ErrPollTimeout) {
				continue // 轮询超时无消息，继续
			}
			mlog.Errorf("[kafka] 读取消息失败: %v", err)
			continue
		}
		d.dispatch(msg)
	}
}

// dispatch 分发消息到对应主题的处理器。同步执行以保持单分区消息顺序，
// 单条消息处理 panic/失败不中断消费循环。
func (d *Kafka) dispatch(msg *Message) {
	d.mu.RLock()
	handler := d.handlers[msg.Topic]
	d.mu.RUnlock()
	if handler == nil {
		mlog.Warnf("[kafka] 主题 %s 无处理器，消息丢弃", msg.Topic)
		return
	}
	safe.Recover(mlog.Fatalf, func() {
		if err := handler(msg.Topic, msg.Value); err != nil {
			mlog.Errorf("[kafka] 处理主题 %s 消息失败: %v", msg.Topic, err)
		}
	})
}

func (d *Kafka) pollTimeout() time.Duration {
	if d.cfg != nil && d.cfg.PollTimeoutMs > 0 {
		return time.Duration(d.cfg.PollTimeoutMs) * time.Millisecond
	}
	return 100 * time.Millisecond
}

// Close 停止消费循环并关闭底层连接。
func (d *Kafka) Close() {
	close(d.exitCh)
	d.wg.Wait()
	d.adapter.Close()
	mlog.Infof("[kafka] 已关闭")
}
