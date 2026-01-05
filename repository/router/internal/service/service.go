package service

import (
	"framework/define"
	"framework/internal/global"
	"framework/library/async"
	"framework/library/mlog"
	"framework/library/structure"
	"framework/library/yaml"
	"framework/repository/router/domain"
	"sync"
	"time"
)

type Service struct {
	ttl        int64
	newFunc    domain.NewFunc    // 创建函数
	filterFunc domain.FilterFunc // 过滤函数
	mutex      sync.RWMutex
	routers    structure.Map2[uint32, uint64, define.IRouter] // 路由表
	exit       chan struct{}                                  // 退出通知
}

func NewService(n domain.NewFunc) *Service {
	return &Service{
		newFunc: n,
		routers: make(structure.Map2[uint32, uint64, define.IRouter]),
		exit:    make(chan struct{}),
	}
}

func (d *Service) Init(cfg *yaml.NodeConfig, filter domain.FilterFunc) {
	d.filterFunc = filter
	d.ttl = cfg.RouterExpire
	async.Go(func() {
		tt := time.NewTicker(12 * time.Second)
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

func (d *Service) Get(idType uint32, id uint64) define.IRouter {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	return nil
}

func (d *Service) GetOrNew(idType uint32, id uint64) define.IRouter {
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item := d.newFunc(idType, id, d.ttl)
	item.Set(global.GetSelf().Type, global.GetSelf().Id)
	d.routers.Put(idType, id, item)
	return item
}

func (d *Service) Remove(idType uint32, id uint64) define.IRouter {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item, _ := d.routers.Del(idType, id)
	return item
}

func (d *Service) refresh(now int64) {
	dels, saves := []define.IRouter{}, []define.IRouter{}
	d.mutex.RLock()
	for _, item := range d.routers {
		if item.IsExpire(now) {
			dels = append(dels, item)
		} else if item.GetStatus() {
			if d.filterFunc != nil && !d.filterFunc(item) {
				saves = append(saves, item)
			}
		}
	}
	d.mutex.RUnlock()

	// 删除过期路由
	d.mutex.Lock()
	for _, item := range dels {
		if vv, ok := d.routers.Del(item.GetIdType(), item.GetId()); ok {
			mlog.Info(0, "删除过期路由记录：%d:%d:%v:%v", vv.GetIdType(), vv.GetId(), ok, vv.GetRouter())
		}
	}
	d.mutex.Unlock()

	// 保存全局路由
	/*
		if d.client != nil && len(saves) > 0 {
			d.save(saves...)
		}
	*/
}
