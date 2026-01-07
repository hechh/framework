package entity

import (
	"context"
	"framework/core"
	"framework/library/async"
	"framework/library/mlog"
	"framework/library/util"

	"path/filepath"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegister struct {
	sync.WaitGroup
	client *clientv3.Client
	prefix string
	exit   chan struct{}
}

func NewEtcdRegister(prefix string, endpoints []string) (ret *EtcdRegister, err error) {
	ret = &EtcdRegister{
		prefix: prefix,
		exit:   make(chan struct{}),
	}
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

func (d *EtcdRegister) Close() {
	close(d.exit)
	d.Wait()
	d.client.Close()
}

func (d *EtcdRegister) Register(key string, val []byte) error {
	// 设置超时时间
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 创建租约
	var lease clientv3.LeaseID
	if err := util.Retry(3, time.Second, func() error {
		rsp, err := d.client.Grant(timeout, core.ETCD_GRANT_TTL)
		if err == nil {
			lease = rsp.ID
		}
		return err
	}); err != nil {
		return err
	}

	// 注册服务
	channel := filepath.Join(d.prefix, key)
	_, err := d.client.Put(ctx, channel, string(val), clientv3.WithLease(lease))
	if err != nil {
		return err
	}

	// 保活
	aliveChan, err := d.client.KeepAlive(ctx, lease)
	if err != nil {
		return err
	}

	d.Add(1)
	async.Go(func() {
		tt := time.NewTicker((core.ETCD_GRANT_TTL / 2) * time.Second)
		defer func() {
			d.Done()
			tt.Stop()
			d.client.Revoke(ctx, lease)
		}()
		for {
			select {
			case _, ok := <-aliveChan:
				if !ok {
					mlog.Error(0, "服务保活失败，重新注册服务中...")
					if err := d.Register(key, val); err != nil {
						mlog.Error(0, "ETCD重新注册服务失败:%v", err)
					} else {
						return
					}
				}
			case <-tt.C:
				if _, err := d.client.TimeToLive(ctx, lease); err != nil {
					mlog.Error(0, "服务保活失败，重新注册服务中...%v", err)
					if err := d.Register(key, val); err != nil {
						mlog.Error(0, "ETCD重新注册服务失败:%v", err)
					} else {
						return
					}
				}
			case <-d.exit:
				return
			}
		}
	})
	return nil
}
