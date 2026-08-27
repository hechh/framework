package kafka_test

import (
	"testing"
	"time"

	"github.com/hechh/framework/pkg/kafka"
	"github.com/hechh/framework/pkg/kafka/adapter/mockkafka"
	"github.com/stretchr/testify/assert"
)

func newTestKafka(t *testing.T) *kafka.Kafka {
	t.Helper()
	k := kafka.NewKafka(mockkafka.NewMock())
	err := k.Init(&kafka.Config{Brokers: "mock", GroupId: "test-group"})
	assert.NoError(t, err)
	return k
}

// TestSubscribeAndConsume 订阅主题后发送消息，消费循环应分发给处理器。
func TestSubscribeAndConsume(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	got := make(chan []byte, 1)
	err := k.Subscribe("topic.test", func(topic string, body []byte) error {
		got <- body
		return nil
	})
	assert.NoError(t, err)

	k.Start()
	assert.NoError(t, k.Send("topic.test", []byte("hello")))

	select {
	case body := <-got:
		assert.Equal(t, "hello", string(body))
	case <-time.After(2 * time.Second):
		t.Fatal("消费循环未收到消息")
	}
}

// TestConsumeLoopOrder 同一主题消息应顺序分发。
func TestConsumeLoopOrder(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	var got []string
	done := make(chan struct{})
	err := k.Subscribe("topic.order", func(topic string, body []byte) error {
		got = append(got, string(body))
		if len(got) == 3 {
			close(done)
		}
		return nil
	})
	assert.NoError(t, err)

	k.Start()
	for _, v := range []string{"1", "2", "3"} {
		assert.NoError(t, k.Send("topic.order", []byte(v)))
	}

	select {
	case <-done:
		assert.Equal(t, []string{"1", "2", "3"}, got)
	case <-time.After(2 * time.Second):
		t.Fatal("消费循环未收到全部消息")
	}
}

// TestSendSync mock 下同步发送与异步等价。
func TestSendSync(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()
	assert.NoError(t, k.SendSync("topic.sync", []byte("sync")))
}

// TestUnhandledTopic 未注册处理器的主题消息不应导致 panic。
func TestUnhandledTopic(t *testing.T) {
	k := newTestKafka(t)
	defer k.Close()

	k.Start()
	assert.NoError(t, k.Send("topic.none", []byte("ignored")))
	time.Sleep(50 * time.Millisecond)
}

// TestComponentInit 组件方式初始化并自动启动消费循环。
func TestComponentInit(t *testing.T) {
	c := &kafka.Component{Object: kafka.NewKafka(mockkafka.NewMock())}
	err := c.Init(map[string]any{
		"kafka": map[string]any{
			"brokers":  "mock",
			"group_id": "test-group",
			"topics":   []string{"topic.cfg"},
		},
	})
	assert.NoError(t, err)
	defer c.Close()
	assert.NotNil(t, kafka.GetObject())
}
