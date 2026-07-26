package router

import (
	"sync"
	"time"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/base/tuple"
	"github.com/hechh/library/mlog"
)

const (
	DEFAULT_ROUTER_TTL = 1800
)

type Router struct {
	self      *packet.Node                             // 当前节点
	mutex     sync.RWMutex                             // 读写锁
	nodeTypes map[uint32]struct{}                      // 支持节点类型
	data      map[tuple.Tuple2[uint32, uint64]]*Entity // 路由数据
	exitCh    chan struct{}                            // 退出通道
}

func NewRouter() *Router {
	return &Router{
		data:   make(map[tuple.Tuple2[uint32, uint64]]*Entity),
		exitCh: make(chan struct{}),
	}
}

// Init 初始化路由表（实现 IComponent 接口）
func (d *Router) Init(cfg *packet.Node, types []uint32) error {
	d.self = cfg
	d.nodeTypes = make(map[uint32]struct{}, len(types))
	for _, t := range types {
		d.nodeTypes[t] = struct{}{}
	}
	go d.run()
	return nil
}

func (d *Router) Close() {
	close(d.exitCh)
}

func (d *Router) run() {
	tt := time.NewTicker(300 * time.Second)
	defer tt.Stop()
	for {
		select {
		case <-d.exitCh:
			return
		case <-tt.C:
			if expired := d.GetExpired(); len(expired) > 0 {
				d.mutex.Lock()
				for _, item := range expired {
					delete(d.data, tuple.T2(item.idType, item.id))
					mlog.Errorf("删除过期缓存路由: %v", item.ToRouteInfo())
				}
				d.mutex.Unlock()
			}
		}
	}
}

func (d *Router) GetExpired() (expired []*Entity) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	now := datetime.NowUnix()
	for _, item := range d.data {
		if now-item.GetAccessTime() >= DEFAULT_ROUTER_TTL {
			expired = append(expired, item)
		}
	}
	return
}

func (d *Router) Get(idType uint32, id uint64) *Entity {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	item, ok := d.data[tuple.T2(idType, id)]
	if ok {
		item.UpdateAccessTime()
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
	key := tuple.T2(idType, id)
	if elem, ok := d.data[key]; ok {
		elem.UpdateAccessTime()
		return elem
	}

	// 创建新路由
	list := NewEntity(idType, id, d.nodeTypes)
	list.Set(d.self.Type, d.self.Id)
	list.UpdateAccessTime()
	d.data[key] = list
	return list
}

func (d *Router) Remove(idType uint32, id uint64) {
	key := tuple.T2(idType, id)
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if elem, ok := d.data[key]; ok {
		delete(d.data, key)
		mlog.Tracef("删除缓存路由: %v", elem.ToRouteInfo())
	}
}
