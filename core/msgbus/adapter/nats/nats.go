package nats

import (
	"fmt"
	"strconv"
	"strings"
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
	Conn   int // 所属连接下标（分片时按 slot 落到对应连接）
}

type Nats struct {
	prefix  string                 // 主题前缀
	conns   []*nats.Conn           // 连接池：conns[0] 广播/回复/发布，分片订阅按 slot 落到对应连接
	subs    [][]*nats.Subscription // 每条连接上的订阅
	entries []*SubscribeEntry      // 所有订阅入口
	workers int                    // 点对点分片数（>=1）
}

func NewNats() *Nats {
	return &Nats{}
}

func (n *Nats) Init(cfg *msgbus.Config) (err error) {
	// 点对点分片并发数：PointWorkers>1 生效，否则 1 条连接
	n.workers = 1
	if cfg.Nats != nil && cfg.Nats.PointWorkers > 1 {
		n.workers = int(cfg.Nats.PointWorkers)
	}

	n.prefix = cfg.Nats.Prefix
	n.conns = make([]*nats.Conn, n.workers)
	n.subs = make([][]*nats.Subscription, n.workers)

	for i := 0; i < n.workers; i++ {
		idx := i
		safe.Retry(3, 3*time.Second, func() error {
			nc, err := nats.Connect(cfg.Nats.Endpoints, n.natsOptions(cfg, idx)...)
			if err != nil {
				mlog.Errorf("[nats] 连接(%d)失败: %v", idx, err)
				return err
			}
			n.conns[idx] = nc
			return nil
		})
		if n.conns[idx] == nil {
			return fmt.Errorf("[nats] 连接(%d)初始化失败", idx)
		}
		mlog.Infof("[nats] 连接(%d)成功: %s", idx, n.conns[idx].ConnectedUrl())
	}
	return nil
}

// Close 优雅关闭：先 Drain 所有订阅再关闭连接
func (n *Nats) Close() {
	for idx, nc := range n.conns {
		if nc == nil {
			continue
		}
		if nc.Status() == nats.CONNECTED {
			if err := nc.Drain(); err != nil {
				mlog.Warnf("[nats] 连接(%d) Drain 失败: %v，直接关闭", idx, err)
				nc.Close()
			} else {
				mlog.Infof("[nats] 连接(%d) 优雅关闭完成", idx)
			}
		} else {
			nc.Close()
		}
	}
	n.conns = nil
	n.subs = nil
	n.entries = nil
}

func (n *Nats) GetRealTopic(topic string) string {
	return n.prefix + "/" + topic
}

// resolveConn 决定订阅落在哪条连接：
//   - 主题尾段是纯数字（分片 slot 主题）时落到对应连接（并发接收，保序由 slot 一致性保证）；
//   - 其余（广播/回复/业务自定义主题）固定 conn[0]。
func (n *Nats) resolveConn(topic string) int {
	if n.workers <= 1 {
		return 0
	}
	segs := strings.Split(topic, "/")
	if len(segs) > 0 {
		if slot, err := strconv.Atoi(segs[len(segs)-1]); err == nil && slot >= 0 {
			return slot % n.workers
		}
	}
	return 0
}

// Subscribe 订阅主题消息，支持断线重连后自动恢复
func (n *Nats) Subscribe(topic string, handle func(*packet.Message)) error {
	entry := &SubscribeEntry{
		Topic:  n.GetRealTopic(topic),
		Handle: handle,
		Conn:   n.resolveConn(topic),
	}
	n.entries = append(n.entries, entry)

	if err := n.subscribeOn(entry); err != nil {
		return err
	}
	return nil
}

func (n *Nats) subscribeOn(entry *SubscribeEntry) error {
	sub, err := n.conns[entry.Conn].Subscribe(entry.Topic, func(msg *nats.Msg) {
		entry.Handle(&packet.Message{Body: msg.Data, Reply: msg.Reply})
	})
	if err != nil {
		return fmt.Errorf("订阅失败: %w", err)
	}
	n.subs[entry.Conn] = append(n.subs[entry.Conn], sub)
	return nil
}

// restoreSubscribe 断线重连后恢复该连接上的所有订阅
func (n *Nats) restoreSubscribe(connIdx int) {
	// 清空旧订阅（避免泄漏）
	for _, sub := range n.subs[connIdx] {
		sub.Unsubscribe()
	}
	n.subs[connIdx] = n.subs[connIdx][:0]

	// 重新订阅该连接的历史主题
	count := 0
	for _, entry := range n.entries {
		if entry.Conn != connIdx {
			continue
		}
		sub, err := n.conns[connIdx].Subscribe(entry.Topic, func(msg *nats.Msg) {
			entry.Handle(&packet.Message{Body: msg.Data, Reply: msg.Reply})
		})
		if err != nil {
			mlog.Errorf("[nats] 恢复订阅失败 topic=%s: %v", entry.Topic, err)
			continue
		}
		n.subs[connIdx] = append(n.subs[connIdx], sub)
		count++
	}
	mlog.Infof("[nats] 连接(%d)全部订阅恢复完成，共恢复 %d 个订阅", connIdx, count)
}

func (n *Nats) natsOptions(cfg *msgbus.Config, connIdx int) []nats.Option {
	opts := []nats.Option{
		nats.DisconnectErrHandler(func(d *nats.Conn, err error) {
			if err != nil {
				mlog.Errorf("[nats] 连接(%d)断开: %s, error: %v", connIdx, d.Opts.Url, err)
			} else {
				mlog.Warnf("[nats] 连接(%d)断开(服务器主动): %s", connIdx, d.Opts.Url)
			}
		}),
		nats.ReconnectHandler(func(d *nats.Conn) {
			mlog.Infof("[nats] 连接(%d)重连成功: %s, 开始恢复订阅", connIdx, d.Opts.Url)
			n.restoreSubscribe(connIdx)
		}),
		nats.ClosedHandler(func(d *nats.Conn) {
			mlog.Warnf("[nats] 连接(%d)已关闭", connIdx)
		}),
		nats.ErrorHandler(func(d *nats.Conn, sub *nats.Subscription, err error) {
			mlog.Errorf("[nats] 连接(%d)异步错误: subject=%s, error=%v", connIdx, sub.Subject, err)
		}),
	}
	ncCfg := cfg.Nats
	if ncCfg.MaxReconnect > 0 {
		opts = append(opts, nats.MaxReconnects(int(ncCfg.MaxReconnect)))
	}
	if ncCfg.ReconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(time.Duration(ncCfg.ReconnectWait)*time.Second))
	}
	if ncCfg.PingInterval > 0 {
		opts = append(opts, nats.PingInterval(time.Duration(ncCfg.PingInterval)*time.Second))
	}
	if ncCfg.DrainTimeout > 0 {
		opts = append(opts, nats.DrainTimeout(time.Duration(ncCfg.DrainTimeout)*time.Second))
	}
	return opts
}

// Publish 发布消息（固定使用连接 0；NATS 服务端按主题路由到对应订阅者）
func (n *Nats) Publish(topic string, body []byte) error {
	return n.conns[0].Publish(n.GetRealTopic(topic), body)
}

// Request 发送同步消息请求
func (n *Nats) Request(topic string, body []byte, cb func([]byte) error) error {
	resp, err := n.conns[0].Request(n.GetRealTopic(topic), body, 3*time.Second)
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}
	return cb(resp.Data)
}

// Response 响应同步消息
func (n *Nats) Response(topic string, body []byte) error {
	err := n.conns[0].Publish(topic, body)
	if err != nil {
		err = fmt.Errorf("publish failed: %w", err)
	}
	return err
}
