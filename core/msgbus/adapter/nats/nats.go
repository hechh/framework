package nats

import (
	"fmt"
	"time"

	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/nats-io/nats.go"
)

type SubscribeEntry struct {
	Topic  string
	Handle func(*packet.Message)
}

type Nats struct {
	client  *nats.Conn           // NATS 连接
	prefix  string               // 主题前缀
	subs    []*nats.Subscription // 所有订阅
	entries []*SubscribeEntry    // 所有订阅入口
}

func NewNats() *Nats {
	return &Nats{}
}

func (n *Nats) Init(cfg *msgbus.Config) (err error) {
	opts := natsOptions(cfg.Nats, n)
	safe.Retry(3, 3*time.Second, func() error {
		n.client, err = nats.Connect(cfg.Nats.Endpoints, opts...)
		return err
	})
	if err == nil {
		n.prefix = cfg.Nats.Prefix
		mlog.Infof("[nats] 连接成功: %s, 服务器: %v", n.client.ConnectedUrl(), n.client.Servers())
	}
	return
}

// Close 优雅关闭 NATS 连接，先 Drain 所有订阅再关闭
func (d *Nats) Close() {
	if d.client.Status() == nats.CONNECTED {
		if err := d.client.Drain(); err != nil {
			mlog.Warnf("[nats] Drain 失败: %v，直接关闭", err)
			d.client.Close()
		} else {
			mlog.Infof("[nats] 优雅关闭完成")
		}
	} else {
		d.client.Close()
		mlog.Infof("[nats] 关闭完成")
	}
	d.client = nil
	d.subs = nil
	d.entries = nil
}

func (d *Nats) GetRealTopic(topic string) string {
	return d.prefix + "/" + topic
}

// Subscribe 订阅主题消息，支持断线重连后自动恢复
func (d *Nats) Subscribe(topic string, handle func(*packet.Message)) error {
	// 记阅入口
	entry := &SubscribeEntry{
		Topic:  d.GetRealTopic(topic),
		Handle: handle,
	}
	d.entries = append(d.entries, entry)

	// 是否已连接
	sub, err := d.client.Subscribe(entry.Topic, func(msg *nats.Msg) {
		entry.Handle(&packet.Message{Body: msg.Data, Reply: msg.Reply})
	})
	if err != nil {
		return fmt.Errorf("subscription failed: %w", err)
	}
	d.subs = append(d.subs, sub)
	return nil
}

// Publish 发布消息到指定主题
func (d *Nats) Publish(topic string, body []byte) error {
	return d.client.Publish(d.GetRealTopic(topic), body)
}

// Request 发送同步消息请求
func (d *Nats) Request(topic string, body []byte, cb func([]byte) error) error {
	resp, err := d.client.Request(d.GetRealTopic(topic), body, 3*time.Second)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}
	return cb(resp.Data)
}

// Response 响应同步消息
func (d *Nats) Response(topic string, body []byte) error {
	err := d.client.Publish(topic, body)
	if err != nil {
		err = fmt.Errorf("publish failed: %w", err)
	}
	return err
}

// restoreSubscribe 断线重连后恢复所有订阅
func (d *Nats) restoreSubscribe() {
	// 清空旧订阅（避免泄漏）
	for _, sub := range d.subs {
		sub.Unsubscribe()
	}
	d.subs = d.subs[:0]

	// 重新订阅所有历史主题
	for _, entry := range d.entries {
		sub, err := d.client.Subscribe(entry.Topic, func(msg *nats.Msg) {
			entry.Handle(&packet.Message{Body: msg.Data, Reply: msg.Reply})
		})
		if err != nil {
			mlog.Errorf("[nats] 恢复订阅失败 topic=%s: %v", entry.Topic, err)
			continue
		}
		d.subs = append(d.subs, sub)
		mlog.Infof("[nats] 恢复订阅成功: %s", entry.Topic)
	}
	mlog.Infof("[nats] 全部订阅恢复完成，共恢复 %d 个订阅", len(d.entries))
}

func natsOptions(cfg *msgbus.NatsConfig, adapter *Nats) []nats.Option {
	opts := []nats.Option{
		nats.DisconnectErrHandler(func(d *nats.Conn, err error) {
			if err != nil {
				mlog.Errorf("[nats] 连接断开: %s, error: %v", d.Opts.Url, err)
			} else {
				mlog.Warnf("[nats] 连接断开(服务器主动): %s", d.Opts.Url)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			mlog.Infof("[nats] 重连成功: %s, 开始恢复所有订阅", nc.ConnectedUrl())
			adapter.restoreSubscribe()
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			mlog.Warnf("[nats] 连接已关闭")
		}),
		nats.DiscoveredServersHandler(func(nc *nats.Conn) {
			mlog.Infof("[nats] 发现新服务器，已知服务器: %v", nc.Servers())
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			mlog.Errorf("[nats] 异步错误: subject=%s, error=%v", sub.Subject, err)
		}),
	}
	if cfg.MaxReconnect > 0 {
		opts = append(opts, nats.MaxReconnects(int(cfg.MaxReconnect)))
	}
	if cfg.ReconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(time.Duration(cfg.ReconnectWait)*time.Second))
	}
	if cfg.PingInterval > 0 {
		opts = append(opts, nats.PingInterval(time.Duration(cfg.PingInterval)*time.Second))
	}
	if cfg.DrainTimeout > 0 {
		opts = append(opts, nats.DrainTimeout(time.Duration(cfg.DrainTimeout)*time.Second))
	}
	return opts
}
