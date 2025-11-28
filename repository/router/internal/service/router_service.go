package service

import (
	"encoding/json"
	"fmt"
	"framework/define"
	"framework/library/mapstruct"
	"framework/library/mlog"
	"framework/library/safe"
	"framework/library/yaml"
	"framework/packet"
	"sync"
	"time"
)

type RouterService struct {
	idType  int32
	mutex   sync.RWMutex
	client  define.IRedis
	newFunc func(int32, uint64) define.IRouter
	routers mapstruct.Map2[int32, uint64, define.IRouter]
	exit    chan struct{}
}

func NewRouterService(f func(int32, uint64) define.IRouter) *RouterService {
	return &RouterService{
		newFunc: f,
		routers: make(mapstruct.Map2[int32, uint64, define.IRouter]),
		exit:    make(chan struct{}),
	}
}

func (d *RouterService) Init(cfg *yaml.NodeConfig, client define.IRedis, idType int32) {
	d.client = client
	d.idType = idType

	safe.Go(func() {
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

func (d *RouterService) Close() {
	close(d.exit)
}

func (d *RouterService) Load(idType int32, id uint64) define.IRouter {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if val, ok := d.routers.Get(idType, id); ok {
		return val
	}
	return nil
}

func (d *RouterService) LoadOrNew(idType int32, id uint64) define.IRouter {
	if item := d.Load(idType, id); item != nil {
		return item
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item := d.newFunc(idType, id)
	d.routers.Set(idType, id, item)
	return item
}

func (d *RouterService) Remove(idType int32, id uint64) define.IRouter {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	item, _ := d.routers.Del(idType, id)
	return item
}

func (d *RouterService) refresh(now int64, expire int64) {
	dels, saves := []define.IRouter{}, []define.IRouter{}
	d.mutex.RLock()
	for _, item := range d.routers {
		if item.IsExpire(now, expire) {
			dels = append(dels, item)
		} else if item.IsChange() && d.idType == item.GetType() {
			saves = append(saves, item)
		}
	}
	d.mutex.RUnlock()

	// 删除过期路由
	d.mutex.Lock()
	for _, item := range dels {
		val, ok := d.routers.Del(item.GetType(), item.GetId())
		mlog.Info(0, "删除过期路由记录：%d:%d:%v:%v", item.GetType(), item.GetId(), ok, val.GetRouter())
	}
	d.mutex.Unlock()

	// 保存全局路由
	if d.client != nil && len(saves) > 0 {
		d.save(saves...)
	}
}

// 路由落地全局路由表
func (d *RouterService) save(rs ...define.IRouter) {
	args := []any{}
	for _, rr := range rs {
		args = append(args, getkey(rr.GetType(), rr.GetId()))
		item := &packet.RouterData{}
		rr.CopyTo(item)
		buf, err := json.Marshal(item)
		if err != nil {
			mlog.Errorf("RouterService更新全局路由失败:%v", err)
			return
		}
		args = append(args, buf)
	}

	// 数据落地
	if err := d.client.MSet(args...); err != nil {
		mlog.Errorf("RouterService更新全局路由失败:%v", err)
		return
	}

	// 修改状态
	for _, rr := range rs {
		rr.Save()
	}
}

func getkey(idType int32, id uint64) string {
	return fmt.Sprintf("router:%d:%d", idType, id)
}
