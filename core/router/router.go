package router

import (
	"sync"

	"github.com/hechh/framework/packet"
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
	return nil
}

func (d *Router) Close() {
	close(d.exitCh)
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
	elem, ok := d.data[key]
	if ok {
		delete(d.data, key)
		mlog.Tracef("删除缓存路由: %v", elem.ToRouteInfo())
	}
	d.mutex.Unlock()
}
