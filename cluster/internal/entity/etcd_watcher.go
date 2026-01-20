package entity

import (
	"context"
	"time"

	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/util"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdWatcher struct {
	prefix string
	client *clientv3.Client
}

func NewEtcdWatcher(prefix string, endpoints []string) (ret *EtcdWatcher, err error) {
	ret = &EtcdWatcher{prefix: prefix}
	util.Retry(3, time.Second, func() error {
		ret.client, err = clientv3.New(clientv3.Config{
			Endpoints:            endpoints,
			DialTimeout:          5 * time.Second,
			DialKeepAliveTime:    30 * time.Second,
			DialKeepAliveTimeout: 3 * time.Second,
			MaxCallSendMsgSize:   10 * 1024 * 1024,
		})
		return err
	})
	return
}

func (d *EtcdWatcher) Close() {
	d.client.Close()
}

func (d *EtcdWatcher) Watch(f func(string, []byte)) error {
	// 设置超时时间
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 同步注册地服务
	rsp, err := d.client.Get(timeout, d.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range rsp.Kvs {
		f(string(kv.Key), kv.Value)
	}

	// 异步监听
	async.Go(func() {
		for {
			wchan := d.client.Watch(ctx, d.prefix, clientv3.WithPrefix())
			if wchan == nil {
				mlog.Error(0, "ETCD(%s)监听失败", d.prefix)
				time.Sleep(time.Second)
				continue
			}
			for rsp := range wchan {
				if rsp.Canceled {
					mlog.Error(0, "Etcd(%s)监听被取消，尝试重新连接", d.prefix)
					break
				}
				if rsp.Err() != nil {
					mlog.Error(0, "Etcd(%s)监听服务出现错误: %v", d.prefix, rsp.Err().Error())
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
	})
	return nil
}
