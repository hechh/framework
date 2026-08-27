package confluent

import (
	"errors"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/hechh/framework/pkg/kafka"
	"github.com/hechh/framework/pkg/mlog"
)

// Confluent confluent-kafka-go（librdkafka）适配器，实现 kafka.IKafka 接口。
// 注意：依赖 CGO + librdkafka，构建时需保证环境具备 librdkafka。
type Confluent struct {
	producer *ckafka.Producer
	consumer *ckafka.Consumer
	cfg      *kafka.Config
}

func NewConfluent() *Confluent {
	return &Confluent{}
}

// Init 初始化生产者与消费者。
func (d *Confluent) Init(cfg *kafka.Config) error {
	d.cfg = cfg

	// 生产者
	producer, err := ckafka.NewProducer(&ckafka.ConfigMap{
		"bootstrap.servers": cfg.Brokers,
	})
	if err != nil {
		return fmt.Errorf("创建 kafka 生产者失败: %w", err)
	}
	d.producer = producer
	// 异步投递结果上报（Send 未显式传 deliveryChan 时投递报告走 Events()）
	go d.deliveryReport()

	// 消费者
	groupId := cfg.GroupId
	if groupId == "" {
		groupId = "kafka-default-consumer"
	}
	cm := ckafka.ConfigMap{
		"bootstrap.servers":  cfg.Brokers,
		"group.id":           groupId,
		"auto.offset.reset":  offsetReset(cfg.AutoOffsetReset),
		"enable.auto.commit": cfg.EnableAutoCommit,
	}
	if cfg.SessionTimeoutMs > 0 {
		cm["session.timeout.ms"] = cfg.SessionTimeoutMs
	}
	consumer, err := ckafka.NewConsumer(&cm)
	if err != nil {
		producer.Close()
		return fmt.Errorf("创建 kafka 消费者失败: %w", err)
	}
	d.consumer = consumer
	return nil
}

// offsetReset 校验 offset 重置策略，非法值回退到 latest。
func offsetReset(v string) string {
	switch v {
	case "earliest", "latest", "none", "error":
		return v
	}
	return "latest"
}

// deliveryReport 消费生产者投递报告通道，记录发送失败与成功。
func (d *Confluent) deliveryReport() {
	for ev := range d.producer.Events() {
		switch e := ev.(type) {
		case *ckafka.Message:
			if e.TopicPartition.Error != nil {
				mlog.Errorf("[kafka] 消息投递失败 topic=%s error=%v", *e.TopicPartition.Topic, e.TopicPartition.Error)
			}
		case ckafka.Error:
			mlog.Errorf("[kafka] 生产者错误: %v", e)
		}
	}
}

// Close 关闭消费者与生产者。
func (d *Confluent) Close() {
	if d.consumer != nil {
		d.consumer.Close()
		d.consumer = nil
	}
	if d.producer != nil {
		d.producer.Flush(3000)
		d.producer.Close()
		d.producer = nil
	}
	mlog.Infof("[kafka] confluent 适配器已关闭")
}

// Subscribe 订阅主题列表。
func (d *Confluent) Subscribe(topics []string) error {
	if d.consumer == nil {
		return errors.New("kafka 消费者未初始化")
	}
	return d.consumer.SubscribeTopics(topics, nil)
}

// Send 异步发送消息到指定主题（不等待投递确认）。
func (d *Confluent) Send(topic string, body []byte) error {
	if d.producer == nil {
		return errors.New("kafka 生产者未初始化")
	}
	return d.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topic},
		Value:          body,
	}, nil)
}

// SendSync 同步发送消息，等待 broker 确认或超时。
func (d *Confluent) SendSync(topic string, body []byte) error {
	if d.producer == nil {
		return errors.New("kafka 生产者未初始化")
	}
	ch := make(chan ckafka.Event, 1)
	if err := d.producer.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topic},
		Value:          body,
	}, ch); err != nil {
		return err
	}
	timeout := 5 * time.Second
	if d.cfg != nil && d.cfg.SendTimeoutMs > 0 {
		timeout = time.Duration(d.cfg.SendTimeoutMs) * time.Millisecond
	}
	select {
	case ev := <-ch:
		if kerr, ok := ev.(ckafka.Error); ok {
			return kerr
		}
		return nil
	case <-time.After(timeout):
		return errors.New("kafka 发送消息超时")
	}
}

// ReadMessage 读取一条消息，超时无消息时返回 kafka.ErrPollTimeout。
func (d *Confluent) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	if d.consumer == nil {
		return nil, errors.New("kafka 消费者未初始化")
	}
	msg, err := d.consumer.ReadMessage(timeout)
	if err != nil {
		if kerr, ok := err.(ckafka.Error); ok && kerr.Code() == ckafka.ErrTimedOut {
			return nil, kafka.ErrPollTimeout
		}
		return nil, err
	}
	return &kafka.Message{
		Topic:  *msg.TopicPartition.Topic,
		Key:    msg.Key,
		Value:  msg.Value,
		Offset: int64(msg.TopicPartition.Offset),
	}, nil
}
