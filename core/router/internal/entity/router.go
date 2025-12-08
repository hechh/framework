package entity

import (
	"encoding/json"
	"fmt"
	"framework/core/define"
	"framework/core/router/domain"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
)

type Router struct {
	idType     uint32
	id         uint64
	data       [define.MAX_NODE_TYPE_COUNT]uint32
	updateTime int64
	change     bool
}

func NewRouter(idType uint32, id uint64) domain.IRouter {
	return &Router{
		idType:     idType,
		id:         id,
		updateTime: time.Now().Unix(),
		data:       [define.MAX_NODE_TYPE_COUNT]uint32{},
	}
}

func (d *Router) GetIdType() uint32 {
	return d.idType
}

func (d *Router) GetId() uint64 {
	return d.id
}

func (d *Router) IsExpire(now int64, expire int64) bool {
	return now-atomic.LoadInt64(&d.updateTime) >= expire
}

func (d *Router) Update() {
	atomic.StoreInt64(&d.updateTime, time.Now().Unix())
}

func (d *Router) Get(nodeType uint32) uint32 {
	return atomic.LoadUint32(&d.data[nodeType-1])
}

func (d *Router) Set(nodeType, nodeId uint32) {
	d.change = !atomic.CompareAndSwapUint32(&d.data[nodeType-1], nodeId, nodeId)
	atomic.StoreUint32(&d.data[nodeType-1], nodeId)
}

func (d *Router) SetRouter(vals ...uint32) {
	for _, val := range vals {
		d.Set(uint32(val&0x1F), uint32(val>>5))
	}
}

func (d *Router) GetRouter() (rets []uint32) {
	for i := 0; i < len(d.data); i++ {
		if val := atomic.LoadUint32(&d.data[i]); val > 0 {
			rets = append(rets, uint32(val<<5)|uint32((i+1)&0x1F))
		}
	}
	return
}

func (d *Router) GetStatus() bool {
	return d.change
}

func (d *Router) SetStatus(val bool) {
	d.change = val
}

func (d *Router) Marshal() (string, error) {
	tmps := map[uint32]uint32{}
	for i := 0; i < len(d.data); i++ {
		if val := atomic.LoadUint32(&d.data[i]); val > 0 {
			tmps[uint32(i)+1] = val
		}
	}
	buf, err := json.Marshal(&tmps)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d|%d|%d|%s", d.idType, d.id, d.updateTime, string(buf)), nil
}

func (d *Router) Unmarshal(str string) error {
	vals := strings.Split(str, "|")

	tmps := map[uint32]uint32{}
	if err := json.Unmarshal([]byte(vals[3]), &tmps); err != nil {
		return err
	}

	d.idType = cast.ToUint32(vals[0])
	d.id = cast.ToUint64(vals[1])
	d.updateTime = cast.ToInt64(vals[2])
	d.change = false

	for nodeType, nodeId := range tmps {
		atomic.StoreUint32(&d.data[nodeType-1], nodeId)
	}
	return nil
}
