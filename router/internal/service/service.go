package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/router/internal/entity"
	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/util"
	"github.com/hechh/library/yaml"
)

type Service struct {
	mutex   sync.RWMutex
	routers util.Map2[uint32, uint64, framework.IRouter] // 路由表
	exit    chan struct{}                                // 退出通知
	ttl     int64                                        // 存活时间
	save    framework.SaveRouterFunc                     // 保存函数
}

func NewService() *Service {
	return &Service{
		routers: make(util.Map2[uint32, uint64, framework.IRouter]),
		exit:    make(chan struct{}),
	}
}

func (d *Service) Init(cfg *yaml.NodeConfig, f framework.SaveRouterFunc) {
	d.ttl = cfg.RouterExpire
	d.save = f
	async.Go(func() {
		tt := time.NewTicker(time.Duration(framework.RouterSyncInterval) * time.Second)
		defer tt.Stop()
		for {
			select {
			case now := <-tt.C:
				d.refresh(now.Unix())
			case <-d.exit:
				return
			}
		}
	})
}

func (d *Service) Close() {
	close(d.exit)
}

func (d *Service) Get(idType uint32, id uint64) framework.IRouter {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	return nil
}

func (d *Service) GetOrNew(idType uint32, id uint64) framework.IRouter {
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item := entity.NewRouter(idType, id, d.ttl)
	item.Set(framework.GetSelfType(), framework.GetSelfId())
	d.routers.Put(idType, id, item)
	return item
}

func (d *Service) Add(str string) error {
	if len(str) <= 0 {
		return nil
	}
	item := entity.NewRouter(0, 0, d.ttl)
	if err := item.Unmarshal(util.StringToBytes(str)); err != nil {
		return err
	}
	if !item.IsExpire(time.Now().Unix()) {
		d.mutex.Lock()
		d.routers.Put(item.GetIdType(), item.GetId(), item)
		d.mutex.Unlock()
	}
	return nil
}

func (d *Service) Remove(idType uint32, id uint64) framework.IRouter {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item, _ := d.routers.Del(idType, id)
	return item
}

func (d *Service) refresh(now int64) {
	dels := []framework.IRouter{}
	tmps := map[string]framework.IRouter{}
	d.mutex.RLock()
	for _, item := range d.routers {
		if item.IsExpire(now) {
			dels = append(dels, item)
		} else if item.GetStatus() && item.GetIdType() == 0 {
			tmps[fmt.Sprintf("%d:%d", item.GetIdType(), item.GetId())] = item
		}
	}
	d.mutex.RUnlock()

	// 删除过期路由
	d.mutex.Lock()
	for _, item := range dels {
		if vv, ok := d.routers.Del(item.GetIdType(), item.GetId()); ok {
			mlog.Info(0, "删除过期路由记录：%d:%d:%v", vv.GetIdType(), vv.GetId(), vv.GetRouter())
		}
	}
	d.mutex.Unlock()

	// 保存全局路由
	if d.save != nil {
		if err := d.save(tmps); err != nil {
			mlog.Errorf("路由表定时保存失败: %v", err)
		}
	}
}
