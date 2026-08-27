//go:build integration

// 集成测试：通过 testcontainers 启动真实 Kafka（KRaft 单节点）broker，
// 端到端验证 segmentio/kafka-go 适配器 + 消费循环的订阅/消费/发送能力。
//
// 运行方式（需 Docker 可用）：
//
//	go test -tags integration ./pkg/kafka/test/ -v
//
// Docker 不可用时测试包自动跳过（静默 PASS）。
package kafka_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/kafka"
	"github.com/hechh/framework/pkg/kafka/adapter/kafkago"
	skafka "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

var (
	testBrokers string // 共享 broker 地址（TestMain 启动一次容器）
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	container, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0")
	if err != nil {
		fmt.Printf("无法启动 Kafka 容器（Docker 不可用？）: %v，跳过集成测试\n", err)
		os.Exit(0)
	}
	brokers, err := container.Brokers(ctx)
	if err != nil {
		fmt.Printf("获取 broker 地址失败: %v\n", err)
		_ = container.Terminate(ctx)
		os.Exit(1)
	}
	testBrokers = brokers[0]
	fmt.Printf("Kafka broker: %s\n", testBrokers)

	// confluent-local 默认不通过 metadata 自动建 topic（segmentio 的
	// AllowAutoTopicCreation 依赖 broker auto.create.topics.enable），
	// 与生产环境"运维预建 topic"一致，这里显式预建测试 topic。
	if err := createTopics(testBrokers, testTopics...); err != nil {
		fmt.Printf("创建测试 topic 失败: %v\n", err)
		_ = container.Terminate(ctx)
		os.Exit(1)
	}

	code := m.Run()
	_ = container.Terminate(ctx)
	os.Exit(code)
}

// testTopics 集成测试使用的全部主题（TestMain 预建）。
var testTopics = []string{
	"it.topic.consume",
	"it.topic.cfg",
	"it.topic.sync",
	"it.topic.a",
	"it.topic.b",
}

// createTopics 通过 plain 连接向 broker 显式创建主题（单分区，副本数 1）。
func createTopics(broker string, topics ...string) error {
	conn, err := skafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("连接 broker 失败: %w", err)
	}
	defer conn.Close()
	cfgs := make([]skafka.TopicConfig, 0, len(topics))
	for _, t := range topics {
		cfgs = append(cfgs, skafka.TopicConfig{Topic: t, NumPartitions: 1, ReplicationFactor: 1})
	}
	return conn.CreateTopics(cfgs...)
}

// uniqGroup 生成唯一消费组 ID，避免重复运行时复用旧提交 offset 导致收不到消息。
func uniqGroup(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// newKafkaGo 创建 segmentio/kafka-go 适配器的 Kafka 封装（earliest 起点保证新消费组能拉到消息）。
func newKafkaGo(t *testing.T, groupId string) *kafka.Kafka {
	t.Helper()
	k := kafka.NewKafka(kafkago.NewKafkaGo())
	require.NoError(t, k.Init(&kafka.Config{
		Brokers:         testBrokers,
		GroupId:         groupId,
		AutoOffsetReset: "earliest",
	}))
	return k
}

// TestIntegrationPublishConsume 发送消息后，消费循环应分发到已注册处理器。
func TestIntegrationPublishConsume(t *testing.T) {
	k := newKafkaGo(t, uniqGroup("it-publish-consume"))
	defer k.Close()

	got := make(chan []byte, 1)
	require.NoError(t, k.Subscribe("it.topic.consume", func(topic string, body []byte) error {
		got <- body
		return nil
	}))
	k.Start()

	require.NoError(t, k.Send("it.topic.consume", []byte("hello-integration")))

	select {
	case body := <-got:
		assert.Equal(t, "hello-integration", string(body))
	case <-time.After(30 * time.Second):
		t.Fatal("30 秒内消费循环未收到消息")
	}
}

// TestIntegrationConfigTopics 通过 Config.Topics 声明订阅（Component 模式），
// 组件初始化后自动启动消费循环，业务侧注册处理器即可收到消息。
func TestIntegrationConfigTopics(t *testing.T) {
	comp := &kafka.Component{Object: kafka.NewKafka(kafkago.NewKafkaGo())}
	err := comp.Init(map[string]any{
		"kafka": map[string]any{
			"brokers":           testBrokers,
			"group_id":          uniqGroup("it-config-topics"),
			"auto_offset_reset": "earliest",
			"topics":            []string{"it.topic.cfg"},
		},
	})
	require.NoError(t, err)
	defer comp.Close()

	got := make(chan []byte, 1)
	require.NoError(t, kafka.GetObject().Subscribe("it.topic.cfg", func(topic string, body []byte) error {
		got <- body
		return nil
	}))

	require.NoError(t, kafka.GetObject().Send("it.topic.cfg", []byte("cfg-message")))

	select {
	case body := <-got:
		assert.Equal(t, "cfg-message", string(body))
	case <-time.After(30 * time.Second):
		t.Fatal("30 秒内组件订阅未收到消息")
	}
}

// TestIntegrationSendSync 同步发送应等待 broker 确认并成功。
func TestIntegrationSendSync(t *testing.T) {
	k := newKafkaGo(t, uniqGroup("it-send-sync"))
	defer k.Close()

	assert.NoError(t, k.SendSync("it.topic.sync", []byte("sync-message")))
}

// TestIntegrationMultipleTopics 一个消费组订阅多个主题，各主题消息均能分发。
func TestIntegrationMultipleTopics(t *testing.T) {
	k := newKafkaGo(t, uniqGroup("it-multi-topics"))
	defer k.Close()

	gotA := make(chan []byte, 1)
	gotB := make(chan []byte, 1)
	require.NoError(t, k.Subscribe("it.topic.a", func(topic string, body []byte) error {
		gotA <- body
		return nil
	}))
	require.NoError(t, k.Subscribe("it.topic.b", func(topic string, body []byte) error {
		gotB <- body
		return nil
	}))
	k.Start()

	require.NoError(t, k.Send("it.topic.a", []byte("msg-a")))
	require.NoError(t, k.Send("it.topic.b", []byte("msg-b")))

	select {
	case body := <-gotA:
		assert.Equal(t, "msg-a", string(body))
	case <-time.After(30 * time.Second):
		t.Fatal("30 秒内未收到主题 a 的消息")
	}
	select {
	case body := <-gotB:
		assert.Equal(t, "msg-b", string(body))
	case <-time.After(30 * time.Second):
		t.Fatal("30 秒内未收到主题 b 的消息")
	}
}
