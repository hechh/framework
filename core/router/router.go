package router

import (
	"sync"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/base/tuple"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
)

type Router struct {
	data      map[tuple.Tuple2[uint32, uint64]]*Entity // 路由数据
	exitCh    chan struct{}                            // 退出通道
	mutex     sync.RWMutex                             // 读写锁
	self      *packet.Node                             // 当前节点
	timer     *timer.Timer                             // 定时器
	nodeTypes []uint32                                 // 支持节点类型
}

func NewRouter() *Router {
	return &Router{
		data:   make(map[tuple.Tuple2[uint32, uint64]]*Entity),
		exitCh: make(chan struct{}),
	}
}

// Init 初始化路由表（实现 IComponent 接口）
func (d *Router) Init(t *timer.Timer, cfg *packet.Node, types []uint32) error {
	d.self = cfg
	d.timer = t
	d.nodeTypes = types
	return nil
}

func (d *Router) Close() {
	close(d.exitCh)
}

func (d *Router) Get(idType uint32, id uint64) *Entity {
	d.mutex.RLock()
	item, ok := d.data[tuple.T2(idType, id)]
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
	key := tuple.T2(idType, id)
	if elem, ok := d.data[key]; ok {
		elem.Refresh(datetime.NowUnixMilli())
		return elem
	}

	// 创建新路由
	list := NewEntity(idType, id, d)
	list.Set(d.self.Type, d.self.Id)
	list.Refresh(datetime.NowUnixMilli())
	d.data[key] = list
	d.timer.Register(list)
	return list
}

func (d *Router) Remove(idType uint32, id uint64) {
	key := tuple.T2(idType, id)
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

func (d *Router) GetSupports() []uint32 {
	return d.nodeTypes
}
