package router

import (
	"sync"

	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/timer"
)

type Router struct {
	data   map[tplutil.Tuple2[uint32, uint64]]*Entity // 路由数据
	exitCh chan struct{}                              // 退出通道
	mutex  sync.RWMutex                               // 读写锁
}

func NewRouter() *Router {
	return &Router{
		data:   make(map[tplutil.Tuple2[uint32, uint64]]*Entity),
		exitCh: make(chan struct{}),
	}
}

func (d *Router) Close() {
	close(d.exitCh)
}

func (d *Router) Get(idType uint32, id uint64) *Entity {
	d.mutex.RLock()
	item, ok := d.data[tplutil.T2(idType, id)]
	d.mutex.RUnlock()
	if ok {
		item.Refresh(datetime.NowUnixMilli())
	}
	return item
}

func (d *Router) GetOrNew(idType uint32, id uint64) *Entity {
	if item := d.Get(idType, id); item != nil {
		return item
	}
	// 慢速路径：写锁创建
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// 双重检查
	key := tplutil.T2(idType, id)
	if elem, ok := d.data[key]; ok {
		elem.Refresh(datetime.NowUnixMilli())
		return elem
	}

	// 创建新路由
	list := NewEntity(idType, id, d)
	list.Set(global.GetSelfNodeType(), global.GetSelfNodeId())
	list.Refresh(datetime.NowUnixMilli())
	d.data[key] = list
	timer.Register(list)
	return list
}

func (d *Router) Remove(idType uint32, id uint64) {
	key := tplutil.T2(idType, id)
	d.mutex.Lock()
	elem, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	d.mutex.Unlock()
	if ok {
		mlog.Tracef("删除缓存路由: %v", elem.ToRouter())
	}
}
