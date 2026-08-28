// Package kafkago 提供基于 segmentio/kafka-go（纯 Go，无 CGO）的 kafka.IKafka 适配器。
package kafkago

import (
	"context"
	"crypto/tls"
	"errors"
	"strings"
	"sync"
	"time"

	skafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"

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
	// 启用 SASL/TLS 认证时注入 Transport（如 Amazon MSK）
	if tr := d.transport(); tr != nil {
		d.writer.Transport = tr
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

// saslMechanism 根据配置构造 SASL 认证机制；未配置或机制未知时返回 nil（匿名访问）。
func (d *KafkaGo) saslMechanism() sasl.Mechanism {
	if d.cfg == nil {
		return nil
	}
	switch strings.ToUpper(d.cfg.SaslMechanism) {
	case "PLAIN":
		return plain.Mechanism{Username: d.cfg.SaslUsername, Password: d.cfg.SaslPassword}
	case "SCRAM-SHA-256":
		m, err := scram.Mechanism(scram.SHA256, d.cfg.SaslUsername, d.cfg.SaslPassword)
		if err != nil {
			mlog.Errorf("[kafka] 构造 SCRAM-SHA-256 机制失败: %v", err)
			return nil
		}
		return m
	case "SCRAM-SHA-512":
		m, err := scram.Mechanism(scram.SHA512, d.cfg.SaslUsername, d.cfg.SaslPassword)
		if err != nil {
			mlog.Errorf("[kafka] 构造 SCRAM-SHA-512 机制失败: %v", err)
			return nil
		}
		return m
	default:
		return nil
	}
}

// useTLS 是否启用 TLS 加密传输（SSL / SASL_SSL 协议）。
func (d *KafkaGo) useTLS() bool {
	if d.cfg == nil {
		return false
	}
	switch strings.ToUpper(d.cfg.SecurityProtocol) {
	case "SSL", "SASL_SSL":
		return true
	default:
		return false
	}
}

// tlsConfig 构造 TLS 配置（内网自签场景允许跳过证书校验）。
func (d *KafkaGo) tlsConfig() *tls.Config {
	return &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: d.cfg != nil && d.cfg.TlsSkipVerify, // #nosec G402 -- 内网自签证书场景允许跳过校验
	}
}

// needAuth 是否启用了 SASL 或 TLS 认证。
func (d *KafkaGo) needAuth() bool {
	return d.saslMechanism() != nil || d.useTLS()
}

// transport 构造生产端 Writer 的传输层（注入 SASL/TLS）；无需认证时返回 nil（保持默认明文）。
func (d *KafkaGo) transport() *skafka.Transport {
	if d.cfg == nil {
		return nil
	}
	mech := d.saslMechanism()
	if mech == nil && !d.useTLS() {
		return nil
	}
	tr := &skafka.Transport{}
	if mech != nil {
		tr.SASL = mech
	}
	if d.useTLS() {
		tr.TLS = d.tlsConfig()
	}
	return tr
}

// dialer 构造消费端 Reader 的拨号器（注入 SASL/TLS）。
func (d *KafkaGo) dialer() *skafka.Dialer {
	dl := &skafka.Dialer{Timeout: 10 * time.Second}
	if mech := d.saslMechanism(); mech != nil {
		dl.SASLMechanism = mech
	}
	if d.useTLS() {
		dl.TLS = d.tlsConfig()
	}
	return dl
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
	// 启用 SASL/TLS 认证时注入 Dialer（如 Amazon MSK）
	if d.needAuth() {
		rc.Dialer = d.dialer()
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
