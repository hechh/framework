package etcdsync

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/mlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdSync struct {
	wg     sync.WaitGroup
	client *clientv3.Client
	exitCh chan struct{}
	prefix string
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
	close(d.exitCh)
	d.wg.Wait()
	d.client.Close()
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

	// 先拉取一次全量配置并回调（初始同步）
	if err := e.Fetch(f); err != nil {
		return nil, err
	}

	watchCh := e.client.Watch(context.Background(), e.prefix, clientv3.WithPrefix())
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
			if !ok || rsp.Canceled {
				mlog.Errorf("配置监听(%s)被取消，尝试重新连接", e.prefix)
				return
			}
			if rsp.Err() != nil {
				mlog.Errorf("配置监听(%s)错误: %v", e.prefix, rsp.Err())
				continue
			}
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
