// Package kafkago 提供基于 segmentio/kafka-go（纯 Go，无 CGO）的 kafka.IKafka 适配器。
package kafkago

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	skafka "github.com/segmentio/kafka-go"

	"github.com/hechh/framework/pkg/kafka"
	"github.com/hechh/framework/pkg/mlog"
)

// KafkaGo segmentio/kafka-go 适配器，实现 kafka.IKafka 接口。
type KafkaGo struct {
	cfg    *kafka.Config
	writer *skafka.Writer

	mu     sync.RWMutex
	topics []string       // 已订阅主题集合（去重）
	reader *skafka.Reader // 消费 Reader（主题集合变化时重建）
}

// NewKafkaGo 创建 segmentio/kafka-go 适配器。
func NewKafkaGo() *KafkaGo {
	return &KafkaGo{}
}

// Init 初始化生产者与消费者配置。
func (d *KafkaGo) Init(cfg *kafka.Config) error {
	brokers := splitBrokers(cfg.Brokers)
	if len(brokers) == 0 {
		return errors.New("kafka brokers 未配置")
	}
	d.cfg = cfg
	// 生产端：多主题共享一个 Writer，消息级别指定 Topic。
	// AllowAutoTopicCreation 依赖 broker auto.create.topics.enable（confluent-local 默认开启）。
	d.writer = &skafka.Writer{
		Addr:                   skafka.TCP(brokers...),
		Balancer:               &skafka.LeastBytes{},
		RequiredAcks:           skafka.RequireAll,
		BatchTimeout:           10 * time.Millisecond,
		AllowAutoTopicCreation: true,
	}
	return nil
}

func splitBrokers(s string) []string {
	var out []string
	for _, b := range strings.Split(s, ",") {
		if b = strings.TrimSpace(b); b != "" {
			out = append(out, b)
		}
	}
	return out
}

// Subscribe 累积订阅主题；主题集合有变化时重建消费 Reader。
func (d *KafkaGo) Subscribe(topics []string) error {
	if len(topics) == 0 {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	changed := false
	for _, t := range topics {
		if t == "" {
			continue
		}
		if !contains(d.topics, t) {
			d.topics = append(d.topics, t)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return d.rebuildReaderLocked()
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// rebuildReaderLocked 根据当前主题集合重建消费 Reader（调用方需持有写锁）。
func (d *KafkaGo) rebuildReaderLocked() error {
	if d.reader != nil {
		_ = d.reader.Close()
		d.reader = nil
	}
	if len(d.topics) == 0 {
		return nil
	}
	groupId := d.cfg.GroupId
	if groupId == "" {
		groupId = "kafka-default-consumer"
	}
	rc := skafka.ReaderConfig{
		Brokers:     splitBrokers(d.cfg.Brokers),
		GroupID:     groupId,
		GroupTopics: append([]string{}, d.topics...),
		StartOffset: startOffset(d.cfg.AutoOffsetReset),
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     100 * time.Millisecond,
	}
	// 自动提交：启用时每 1s 提交一次；关闭时禁用（与 confluent enable.auto.commit 语义一致）
	if d.cfg.EnableAutoCommit {
		rc.CommitInterval = time.Second
	} else {
		rc.CommitInterval = 0
	}
	if d.cfg.SessionTimeoutMs > 0 {
		rc.SessionTimeout = time.Duration(d.cfg.SessionTimeoutMs) * time.Millisecond
	}
	d.reader = skafka.NewReader(rc)
	return nil
}

func startOffset(v string) int64 {
	switch v {
	case "earliest":
		return skafka.FirstOffset
	default:
		return skafka.LastOffset
	}
}

// Send 异步发送消息（goroutine 中同步写入，不阻塞调用方）。
func (d *KafkaGo) Send(topic string, body []byte) error {
	d.mu.RLock()
	w := d.writer
	d.mu.RUnlock()
	if w == nil {
		return errors.New("kafka 生产者未初始化")
	}
	payload := make([]byte, len(body))
	copy(payload, body)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), d.sendTimeout())
		defer cancel()
		if err := w.WriteMessages(ctx, skafka.Message{Topic: topic, Value: payload}); err != nil {
			mlog.Errorf("[kafka] 异步发送失败 topic=%s error=%v", topic, err)
		}
	}()
	return nil
}

// SendSync 同步发送消息，等待 broker 确认或超时。
func (d *KafkaGo) SendSync(topic string, body []byte) error {
	d.mu.RLock()
	w := d.writer
	d.mu.RUnlock()
	if w == nil {
		return errors.New("kafka 生产者未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.sendTimeout())
	defer cancel()
	return w.WriteMessages(ctx, skafka.Message{Topic: topic, Value: body})
}

func (d *KafkaGo) sendTimeout() time.Duration {
	if d.cfg != nil && d.cfg.SendTimeoutMs > 0 {
		return time.Duration(d.cfg.SendTimeoutMs) * time.Millisecond
	}
	return 5 * time.Second
}

// ReadMessage 读取一条消息；超时无消息时返回 kafka.ErrPollTimeout。
func (d *KafkaGo) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	d.mu.RLock()
	r := d.reader
	d.mu.RUnlock()
	if r == nil {
		return nil, errors.New("kafka 消费者未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	msg, err := r.FetchMessage(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, kafka.ErrPollTimeout
		}
		return nil, err
	}
	return &kafka.Message{
		Topic:  msg.Topic,
		Key:    msg.Key,
		Value:  msg.Value,
		Offset: msg.Offset,
	}, nil
}

// Close 关闭消费 Reader 与生产者。
func (d *KafkaGo) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.reader != nil {
		_ = d.reader.Close()
		d.reader = nil
	}
	if d.writer != nil {
		_ = d.writer.Close()
		d.writer = nil
	}
	mlog.Infof("[kafka] kafkago 适配器已关闭")
}
