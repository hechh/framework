package discovery

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/hechh/framework/kernel/cluster"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/mlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Etcd struct {
	wg     sync.WaitGroup
	client *clientv3.Client
	exitCh chan struct{}
	prefix string
	ttl    int64
}

func NewEtcd() *Etcd {
	return &Etcd{exitCh: make(chan struct{})}
}

func (e *Etcd) Init(cfg *cluster.Config) error {
	e.prefix = cfg.Etcd.Prefix
	e.ttl = cfg.Etcd.KeepAlive
	var err error
	return safe.Retry(3, 3*time.Second, func() error {
		e.client, err = clientv3.New(clientv3.Config{
			Endpoints:            cfg.Etcd.Endpoints,
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    30 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			MaxCallSendMsgSize:   10 * 1024 * 1024,
		})
		if err != nil {
			err = fmt.Errorf("discovery init failed: %w", err)
		}
		return err
	})
}

func (e *Etcd) Close() {
	close(e.exitCh)
	e.wg.Wait()
	e.client.Close()
}

func (e *Etcd) Register(key string, body []byte) error {
	// 首次注册
	lease, aliveCh, err := e.register(key, body)
	if err != nil {
		return err
	}

	e.wg.Add(1)
	safe.SafeGo(mlog.Fatalf, func() {
		defer e.wg.Done()
		for {
			// 维持租约，阻塞直到保活失败或退出
			e.keepAlive(key, lease, aliveCh)

			// 检查退出信号
			select {
			case <-e.exitCh:
				return
			default:
			}

			// 保活失败，重新注册
			var regErr error
			if lease, aliveCh, regErr = e.register(key, body); regErr != nil {
				mlog.Errorf("服务(%s)重新注册失败: %v", key, regErr)
				time.Sleep(3 * time.Second)
			}
		}
	})
	return nil
}

// register 注册节点
func (e *Etcd) register(key string, body []byte) (clientv3.LeaseID, <-chan *clientv3.LeaseKeepAliveResponse, error) {
	var err error
	var rsp *clientv3.LeaseGrantResponse

	// 申请租约
	safe.Retry(3, 3*time.Second, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if rsp, err = e.client.Grant(ctx, e.ttl); err != nil {
			err = fmt.Errorf("grant failed: %w", err)
		}
		return err
	})
	if err != nil {
		return 0, nil, err
	}

	// 注册节点信息
	topic := path.Join(e.prefix, key)
	_, err = e.client.Put(context.Background(), topic, string(body), clientv3.WithLease(rsp.ID))
	if err != nil {
		return 0, nil, fmt.Errorf("put failed: %w", err)
	}

	// 保持租约
	aliveCh, err := e.client.KeepAlive(context.Background(), rsp.ID)
	if err != nil {
		return 0, nil, fmt.Errorf("keepalive failed: %w", err)
	}
	return rsp.ID, aliveCh, nil
}

func (e *Etcd) keepAlive(key string, lease clientv3.LeaseID, aliveCh <-chan *clientv3.LeaseKeepAliveResponse) {
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		e.client.Revoke(ctx, lease)
		cancel()
	}()

	for {
		select {
		case _, ok := <-aliveCh:
			if !ok {
				// etcd KeepAlive 通道关闭 = 租约丢失或连接断开
				mlog.Errorf("服务(%s)保活通道关闭，准备重新注册...", key)
				return
			}
		case <-e.exitCh:
			return
		}
	}
}

func (e *Etcd) Watch(f func(string, []byte)) error {
	watchCh, err := e.watch(f)
	if err != nil {
		return err
	}

	e.wg.Add(1)
	safe.SafeGo(mlog.Fatalf, func() {
		defer e.wg.Done()
		for {
			// 监听节点变化，阻塞直到 watch 被取消或出错
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
				mlog.Errorf("服务发现(%s)重新注册失败: %v", e.prefix, watchErr)
				time.Sleep(3 * time.Second)
			}
		}
	})

	return nil
}

func (e *Etcd) watch(f func(string, []byte)) (clientv3.WatchChan, error) {
	// 监听节点变化
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rsp, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("get failed: %w", err)
	}

	for _, ev := range rsp.Kvs {
		f(string(ev.Key), ev.Value)
	}

	watchCh := e.client.Watch(context.Background(), e.prefix, clientv3.WithPrefix())
	if watchCh == nil {
		return nil, fmt.Errorf("watch channel is nil")
	}
	return watchCh, nil
}

func (e *Etcd) monitor(watchCh clientv3.WatchChan, f func(string, []byte)) {
	for {
		select {
		case <-e.exitCh:
			return
		case rsp, ok := <-watchCh:
			if !ok || rsp.Canceled {
				mlog.Errorf("服务发现(%s)监听被取消，尝试重新连接", e.prefix)
				return
			}
			if rsp.Err() != nil {
				mlog.Errorf("服务发现(%s)监听错误: %v", e.prefix, rsp.Err())
				continue
			}
			// 处理变更事件
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
