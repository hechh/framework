package router

import (
	"sync"

	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/timer"
)

type shardData struct {
	data  map[tplutil.Tuple2[uint32, uint64]]*Entity // 路由数据
	mutex sync.RWMutex                               // 读写锁
}

type Router struct {
	shards []*shardData
}

func NewRouter() *Router {
	shards := make([]*shardData, 0, 256)
	for range 256 {
		shards = append(shards, &shardData{
			data: make(map[tplutil.Tuple2[uint32, uint64]]*Entity),
		})
	}
	return &Router{
		shards: shards,
	}
}

func (d *Router) Get(idType uint32, id uint64) *Entity {
	shard := d.shards[id%256]
	shard.mutex.RLock()
	item, ok := shard.data[tplutil.T2(idType, id)]
	shard.mutex.RUnlock()
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
	shard := d.shards[id%256]
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	// 双重检查
	key := tplutil.T2(idType, id)
	if elem, ok := shard.data[key]; ok {
		elem.Refresh(datetime.NowUnixMilli())
		return elem
	}

	// 创建新路由
	list := NewEntity(idType, id, d)
	list.Set(global.GetSelfNodeType(), global.GetSelfNodeId())
	list.Refresh(datetime.NowUnixMilli())
	shard.data[key] = list
	timer.Register(list)
	return list
}

func (d *Router) Remove(idType uint32, id uint64) {
	shard := d.shards[id%256]
	key := tplutil.T2(idType, id)
	shard.mutex.Lock()
	elem, ok := shard.data[key]
	if ok {
		delete(shard.data, key)
	}
	shard.mutex.Unlock()
	if ok {
		mlog.Tracef("删除缓存路由: %v", elem.ToRouter())
	}
}
