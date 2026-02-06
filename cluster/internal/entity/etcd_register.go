package entity

import (
	"context"
	"path"

	"github.com/hechh/framework"
	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/util"

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
		rsp, err := d.client.Grant(timeout, framework.ETCD_GRANT_TTL)
		if err == nil {
			lease = rsp.ID
		}
		return err
	}); err != nil {
		return err
	}

	// 注册服务
	channel := path.Join(d.prefix, key)
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
		tt := time.NewTicker((framework.ETCD_GRANT_TTL / 2) * time.Second)
		defer func() {
			d.Done()
			tt.Stop()
			d.client.Revoke(ctx, lease)
		}()
		for {
			select {
			case _, ok := <-aliveChan:
				if !ok {
					mlog.Errorf("服务保活失败，重新注册服务中...")
					if err := d.Register(key, val); err != nil {
						mlog.Errorf("ETCD重新注册服务失败:%v", err)
					} else {
						return
					}
				}
			case <-tt.C:
				if _, err := d.client.TimeToLive(ctx, lease); err != nil {
					mlog.Errorf("服务保活失败，重新注册服务中...%v", err)
					if err := d.Register(key, val); err != nil {
						mlog.Errorf("ETCD重新注册服务失败:%v", err)
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

/*
type EtcdRegister struct {
	sync.WaitGroup
	client *clientv3.Client
	prefix string
	exit   chan struct{}
}

func (d *EtcdRegister) Init(cfg *yaml.EtcdConfig) error {
	d.prefix = cfg.Topic
	d.exit = make(chan struct{})
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

func (d *EtcdRegister) Close() {
	close(d.exit)
	d.Wait()
	d.client.Close()
}

// updateValue 定时更新指定key的值[1](@ref)
func (d *EtcdRegister) updateValue(fullPath string, originalVal []byte, lease clientv3.LeaseID) error {
	// 可以在这里实现值更新逻辑，例如添加时间戳或版本信息
	updatedVal := fmt.Sprintf("%s_updated_%d", string(originalVal), time.Now().Unix())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := d.client.Put(ctx, fullPath, updatedVal, clientv3.WithLease(lease))
	if err != nil {
		mlog.Errorf("更新键值失败: %v", err)
		return err
	}
	mlog.Infof("键值更新成功: %s -> %s", fullPath, updatedVal)
	return nil
}

func (d *EtcdRegister) Register(key string, val []byte) error {
	ctx := context.Background()
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 创建租约[3,6](@ref)
	var lease clientv3.LeaseID
	if err := util.Retry(3, time.Second, func() error {
		rsp, err := d.client.Grant(timeout, framework.ETCD_GRANT_TTL)
		if err == nil {
			lease = rsp.ID
		}
		return err
	}); err != nil {
		return err
	}

	// 注册服务[5,8](@ref)
	channel := path.Join(d.prefix, key)
	_, err := d.client.Put(ctx, channel, string(val), clientv3.WithLease(lease))
	if err != nil {
		return err
	}

	// 启动保活和更新goroutine[1,7](@ref)
	d.Add(1)
	async.Go(func() {
		defer func() {
			d.Done()
			// 退出时撤销租约[7](@ref)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			d.client.Revoke(ctx, lease)
			mlog.Infof("服务注销完成: %s", channel)
		}()

		// 主保活通道[6](@ref)
		aliveChan, err := d.client.KeepAlive(ctx, lease)
		if err != nil {
			mlog.Errorf("创建保活通道失败: %v", err)
			return
		}

		// 定时器设置[1](@ref)
		valueUpdateTicker := time.NewTicker(framework.ETCD_UPDATE_INTERVAL * time.Second)
		healthCheckTicker := time.NewTicker((framework.ETCD_GRANT_TTL / 3) * time.Second)
		defer func() {
			valueUpdateTicker.Stop()
			healthCheckTicker.Stop()
		}()

		retryCount := 0
		const maxRetryCount = 3

		for {
			select {
			case resp, ok := <-aliveChan:
				if !ok {
					mlog.Warn("保活通道关闭，进行健康检查...")
					// 通道关闭时通过健康检查验证状态
					continue
				}
				if resp != nil {
					retryCount = 0 // 重置重试计数
					mlog.Debugf("保活成功: TTL=%d", resp.TTL)
				}

			case <-valueUpdateTicker.C:
				// 定时更新键值[1](@ref)
				if err := d.updateValue(channel, val, lease); err != nil && retryCount < maxRetryCount {
					retryCount++
					mlog.Warnf("值更新失败，重试 %d/%d", retryCount, maxRetryCount)
				}

			case <-healthCheckTicker.C:
				// 备用健康检查[7](@ref)
				ttlResp, err := d.client.TimeToLive(ctx, lease)
				if err != nil || ttlResp.TTL <= 0 {
					mlog.Warnf("健康检查失败: %v, 尝试重新注册", err)
					if retryCount < maxRetryCount {
						retryCount++
						// 有限次重试注册[2](@ref)
						if err := d.retryRegister(key, val, retryCount, maxRetryCount); err == nil {
							return
						}
					} else {
						mlog.Error("超过最大重试次数，停止服务")
						return
					}
				} else {
					retryCount = 0 // 重置重试计数
				}

			case <-d.exit:
				mlog.Info("接收到退出信号，停止服务保活")
				return
			}
		}
	})
	return nil
}

// retryRegister 有限次重试注册[2](@ref)
func (d *EtcdRegister) retryRegister(key string, val []byte, retryCount, maxRetryCount int) error {
	mlog.Infof("尝试重新注册服务 (%d/%d)", retryCount, maxRetryCount)

	// 短暂延迟后重试
	time.Sleep(time.Duration(retryCount) * time.Second)

	if err := d.Register(key, val); err != nil {
		mlog.Errorf("重新注册失败: %v", err)
		return err
	}
	return nil
}
*/
