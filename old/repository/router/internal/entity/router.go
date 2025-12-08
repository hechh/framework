package entity

import (
	"framework/define"
	"framework/internal/global"
	"framework/packet"
	"sync/atomic"
	"time"
)

type Router struct {
	idType     int32
	id         uint64
	updateTime int64
	change     bool
	data       [define.MAX_NODE_TYPE_COUNT]int32
}

func NewRouter(idType int32, id uint64) define.IRouter {
	ret := &Router{
		idType:     idType,
		id:         id,
		updateTime: time.Now().Unix(),
		data:       [define.MAX_NODE_TYPE_COUNT]int32{},
		change:     true,
	}
	ret.Set(global.GetSelfType(), global.GetSelfId())
	return ret
}

func (d *Router) GetType() int32 {
	return d.idType
}

func (d *Router) GetId() uint64 {
	return d.id
}

func (d *Router) Get(nodeType int32) int32 {
	return atomic.LoadInt32(&d.data[nodeType-1])
}

func (d *Router) Set(nodeType int32, nodeId int32) {
	d.change = !atomic.CompareAndSwapInt32(&d.data[nodeType-1], nodeId, nodeId)
	atomic.StoreInt32(&d.data[nodeType-1], nodeId)
}

func (d *Router) IsExpire(now int64, expire int64) bool {
	return now-atomic.LoadInt64(&d.updateTime) >= expire
}

func (d *Router) Update() {
	atomic.StoreInt64(&d.updateTime, time.Now().Unix())
}

func (d *Router) SetRouter(vals ...uint32) {
	for _, val := range vals {
		d.Set(int32(val&0x1F), int32(val>>5))
	}
}

func (d *Router) GetRouter() (rets []uint32) {
	for i := 0; i < len(d.data); i++ {
		if val := atomic.LoadInt32(&d.data[i]); val > 0 {
			rets = append(rets, uint32(val<<5)|uint32((i+1)&0x1F))
		}
	}
	return
}

func (d *Router) IsChange() bool {
	return d.change && d.idType == 0
}

func (d *Router) Change() {
	d.change = true
}

func (d *Router) Save() {
	d.change = false
}

func (d *Router) CopyTo(rsp any) {
	data, ok := rsp.(*packet.RouterData)
	if !ok || data == nil {
		return
	}
	data.Id = d.id
	data.IdType = d.idType
	data.UpdateTime = d.updateTime
	for i := 0; i < len(d.data); i++ {
		if val := atomic.LoadInt32(&d.data[i]); val > 0 {
			data.Routers = append(data.Routers, &packet.Node{
				Type: int32(i) + 1,
				Id:   val,
			})
		}
	}
}
