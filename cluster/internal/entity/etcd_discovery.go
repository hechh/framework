package entity

import (
	"context"
	"path"
	"time"

	"github.com/hechh/library/uerror"
	"github.com/hechh/library/util"
	"github.com/hechh/library/yaml"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdDiscovery struct {
	client *clientv3.Client
	lease  clientv3.LeaseID
	prefix string
	ttl    int64
}

func (d *EtcdDiscovery) Init(cfg *yaml.EtcdConfig) error {
	d.prefix = cfg.Topic
	d.ttl = cfg.Expire
	etcdCfg := clientv3.Config{
		Endpoints:            cfg.Endpoints,
		DialTimeout:          5 * time.Second,
		DialKeepAliveTime:    30 * time.Second,
		DialKeepAliveTimeout: 3 * time.Second,
		MaxCallSendMsgSize:   10 * 1024 * 1024,
	}
	return util.Retry(3, time.Second, func() error {
		cli, err := clientv3.New(etcdCfg)
		if err == nil {
			d.client = cli
		}
		return err
	})
}

func (d *EtcdDiscovery) Close() {
	d.client.Close()
}

func (d *EtcdDiscovery) KeepAlive() error {
	// 租赁
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if rsp, err := d.client.Grant(ctx, d.ttl); err == nil {
		d.lease = rsp.ID
	} else {
		return err
	}

	// 主保活通道
	aliveChan, err := d.client.KeepAlive(context.Background(), d.lease)
	if err != nil {
		return err
	}

	// 定时器设置
	healthCheckTicker := time.NewTicker((time.Duration(d.ttl) / 3) * time.Second)
	defer healthCheckTicker.Stop()
	for {
		select {
		case _, ok := <-aliveChan:
			if !ok {
				return uerror.Err(-1, "Etcd保活通道关闭，进行健康检查...")
			}
		case <-healthCheckTicker.C:
			if _, err := d.client.TimeToLive(context.Background(), d.lease); err != nil {
				return err
			}
		}
	}
}

func (d *EtcdDiscovery) Put(key string, body []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := d.client.Put(ctx, path.Join(d.prefix, key), util.BytesToString(body), clientv3.WithLease(d.lease))
	return err
}

func (d *EtcdDiscovery) Get(f func(string, []byte)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 读取kv对
	rsp, err := d.client.Get(ctx, d.prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range rsp.Kvs {
		f(string(kv.Key), kv.Value)
	}
	return nil
}
