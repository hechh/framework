package etcdsync

import (
	"context"
	"fmt"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/mlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdSync struct {
	wg        sync.WaitGroup
	client    *clientv3.Client
	exitCh    chan struct{}
	closeOnce sync.Once // Close 幂等：避免 close(exitCh) 二次执行 panic
	prefix    string

	// lastRev 已处理到的 etcd revision：Fetch 后置为快照 revision，监听中随每个响应推进。
	// 重连时从 lastRev+1 续传，使「快照与建流之间」以及「断线期间」的变更都能被重放，
	// 避免配置静默漏更（etcd watch 不带 WithRev 时只收建流之后的新事件）。
	lastRev atomic.Uint64
}

func NewEtcdSync() *EtcdSync {
	return &EtcdSync{
		exitCh: make(chan struct{}),
	}
}

func (d *EtcdSync) Init(cfg *fwatcher.Config) error {
	d.prefix = cfg.Etcd.Prefix
	var err error
	return safe.Retry(3, 3*time.Second, func() error {
		d.client, err = clientv3.New(clientv3.Config{
			Endpoints:            cfg.Etcd.Endpoints,
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    30 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			MaxCallSendMsgSize:   10 * 1024 * 1024,
		})
		return err
	})
}

func (d *EtcdSync) Close() {
	// 幂等：初始化失败路径与组件 Close 都可能调用，重复 close 会 panic
	d.closeOnce.Do(func() {
		close(d.exitCh)
		d.wg.Wait()
		d.client.Close()
	})
}

func (d *EtcdSync) Put(sheet string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topic := path.Join(d.prefix, sheet)
	_, err := d.client.Put(ctx, topic, string(body))
	if err != nil {
		return err
	}
	mlog.Tracef("上传配置(%s)成功", topic)
	return nil
}

// Clear 清空 prefix 下所有 kv（用于全量同步前清理残留配置）。
func (d *EtcdSync) Clear() error {
	// 防御：prefix 为空时拒绝清空，避免误删整个 etcd（含集群节点注册）
	if d.prefix == "" {
		return fmt.Errorf("etcd prefix 为空，拒绝清空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := d.client.Delete(ctx, d.prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("etcd clear: %w", err)
	}
	mlog.Tracef("清空配置(%s)成功", d.prefix)
	return nil
}

func (d *EtcdSync) Update(sheet string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topic := path.Join(d.prefix, sheet)
	txn := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Version(topic), ">", 0)).
		Then(clientv3.OpPut(topic, string(body))).
		Else(clientv3.OpGet(topic))

	txnResp, err := txn.Commit()
	if err != nil {
		return err
	}
	if !txnResp.Succeeded {
		return fmt.Errorf("config %q not found, cannot update", topic)
	}
	mlog.Tracef("更新配置(%s)成功", topic)
	return nil
}

func (d *EtcdSync) Delete(sheet string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	topic := path.Join(d.prefix, sheet)
	_, err := d.client.Delete(ctx, topic)
	if err != nil {
		return fmt.Errorf("etcd delete: %w", err)
	}
	mlog.Tracef("删除配置(%s)成功", topic)
	return nil
}

// Fetch 拉取 prefix 下所有 kv 并逐个回调（用于消费者模式初始化时从 etcd 同步全量配置到本地）。
func (d *EtcdSync) Fetch(f func(string, []byte)) error {
	// 防御：prefix 为空时拒绝拉取
	if d.prefix == "" {
		return fmt.Errorf("etcd prefix 为空，拒绝拉取")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rsp, err := d.client.Get(ctx, d.prefix, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("etcd fetch: %w", err)
	}

	for _, ev := range rsp.Kvs {
		f(string(ev.Key), ev.Value)
	}

	// 记录快照 revision 作为续传基线：后续 watch 从该 revision 之后开始，
	// 保证「Fetch 快照」与「建流」之间发生的变更也能被 etcd 重放，不会漏更。
	d.lastRev.Store(uint64(rsp.Header.Revision))
	return nil
}

func (e *EtcdSync) Watch(f func(string, []byte)) error {
	watchCh, err := e.watch(f)
	if err != nil {
		return err
	}

	e.wg.Add(1)
	safe.SafeGo(mlog.Fatalf, func() {
		defer e.wg.Done()
		for {
			// 监听配置变化，阻塞直到 watch 被取消或出错
			e.monitor(watchCh, f)

			// 检查退出信号
			select {
			case <-e.exitCh:
				return
			default:
			}

			// watch 断开，重新建立
			var watchErr error
			if watchCh, watchErr = e.watch(f); watchErr != nil {
				mlog.Errorf("配置监听(%s)重新注册失败: %v", e.prefix, watchErr)
				time.Sleep(3 * time.Second)
			}
		}
	})

	return nil
}

func (e *EtcdSync) watch(f func(string, []byte)) (clientv3.WatchChan, error) {
	// 防御：prefix 为空时拒绝监听，避免拉取/监听整个 etcd（含集群节点数据）
	if e.prefix == "" {
		return nil, fmt.Errorf("etcd prefix 为空，拒绝监听")
	}

	// 按 revision 续传：从 lastRev+1 开始监听，使 etcd 重放续传点之后的变更（不丢不重）。
	// lastRev==0 表示尚无基线（未 Fetch 过），此时从当前 revision 开始监听。
	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if rev := e.lastRev.Load(); rev > 0 {
		opts = append(opts, clientv3.WithRev(int64(rev+1)))
	}

	watchCh := e.client.Watch(context.Background(), e.prefix, opts...)
	if watchCh == nil {
		return nil, fmt.Errorf("watch channel is nil")
	}
	return watchCh, nil
}

func (e *EtcdSync) monitor(watchCh clientv3.WatchChan, f func(string, []byte)) {
	for {
		select {
		case <-e.exitCh:
			return
		case rsp, ok := <-watchCh:
			if !ok {
				// 通道关闭（如 etcd 重启导致 gRPC 流断开）：由 Watch 循环按 lastRev+1 续传重连
				mlog.Errorf("配置监听(%s)通道关闭，尝试重连并续传", e.prefix)
				return
			}
			// 续传点已被压缩：etcd 无法重放断线期间的变更，回退全量拉取重建基线
			if rsp.CompactRevision != 0 {
				mlog.Warnf("配置监听(%s)续传点已被压缩(compact_revision=%d)，回退全量拉取重建基线",
					e.prefix, rsp.CompactRevision)
				if err := e.Fetch(f); err != nil {
					mlog.Errorf("配置监听(%s)回退全量拉取失败: %v", e.prefix, err)
				}
				return
			}
			if rsp.Canceled {
				// 服务端取消（非压缩，如续传 revision 超出可服务范围）：续传点不可用，
				// 回退全量拉取重建基线，避免陷入「续传即被取消」的重连空转
				mlog.Errorf("配置监听(%s)被服务端取消，回退全量拉取重建基线", e.prefix)
				if err := e.Fetch(f); err != nil {
					mlog.Errorf("配置监听(%s)回退全量拉取失败: %v", e.prefix, err)
				}
				return
			}
			if err := rsp.Err(); err != nil {
				mlog.Errorf("配置监听(%s)错误: %v", e.prefix, err)
				return
			}

			// 推进续传点：本批事件均已包含在该 revision 内，重连从其后继续
			e.lastRev.Store(uint64(rsp.Header.Revision))
			for _, event := range rsp.Events {
				switch event.Type {
				case clientv3.EventTypePut:
					f(string(event.Kv.Key), event.Kv.Value)
				case clientv3.EventTypeDelete:
					f(string(event.Kv.Key), nil)
				}
			}
		}
	}
}
