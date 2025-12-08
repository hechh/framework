package service

import (
	"framework/core/router/domain"
	"framework/library/async"
	"framework/library/mapstruct"
	"framework/library/mlog"
	"framework/library/yaml"
	"time"
)

type Service struct {
	newFunc    domain.NewFunc                                   // 创建函数
	filterFunc domain.FilterFunc                                // 过滤函数
	routers    *mapstruct.Map2S[uint32, uint64, domain.IRouter] // 路由表
	exit       chan struct{}                                    // 退出通知
}

func NewService(n domain.NewFunc) *Service {
	return &Service{
		newFunc: n,
		routers: mapstruct.NewMap2S[uint32, uint64, domain.IRouter](),
		exit:    make(chan struct{}),
	}
}

func (d *Service) Init(cfg *yaml.NodeConfig, filter domain.FilterFunc) {
	d.filterFunc = filter
	async.Go(func() {
		tt := time.NewTicker(12 * time.Second)
		defer tt.Stop()
		for {
			select {
			case now := <-tt.C:
				d.refresh(now.Unix(), cfg.RouterExpire)
			case <-d.exit:
				return
			}
		}
	})
}

func (d *Service) Close() {
	close(d.exit)
}

func (d *Service) Get(idType uint32, id uint64) domain.IRouter {
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	return nil
}

func (d *Service) GetOrNew(idType uint32, id uint64) domain.IRouter {
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	item := d.newFunc(idType, id)
	d.routers.Set(idType, id, item)
	return item
}

func (d *Service) Remove(idType uint32, id uint64) domain.IRouter {
	item, _ := d.routers.Del(idType, id)
	return item
}

func (d *Service) refresh(now int64, expire int64) {
	dels, saves := []domain.IRouter{}, []domain.IRouter{}
	d.routers.Walk(func(item domain.IRouter) bool {
		if item.IsExpire(now, expire) {
			dels = append(dels, item)
		} else if item.GetStatus() {
			if d.filterFunc != nil && !d.filterFunc(item) {
				saves = append(saves, item)
			} else {
				saves = append(saves, item)
			}
		}
		return true
	})

	// 删除过期路由
	for _, item := range dels {
		vv, ok := d.routers.Del(item.GetIdType(), item.GetId())
		if ok {
			mlog.Info(0, "删除过期路由记录：%d:%d:%v:%v", vv.GetIdType(), vv.GetId(), ok, vv.GetRouter())
		}
	}

	// 保存全局路由
	/*
		if d.client != nil && len(saves) > 0 {
			d.save(saves...)
		}
	*/
}
