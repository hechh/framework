package context

import (
	"framework/internal/global"
	"strings"
	"sync/atomic"
)

type Common struct {
	actorName    string
	funcName     string
	depth        uint32
	routerId     uint64
	idType       int32
	id           uint64
	dstNodeType  int32
	dstActorFunc string
	dstActorId   uint64
	cbNodeType   int32
	cbNodeId     int32
	cbActorFunc  string
	cbActorId    uint64
}

func NewCommon(actorFunc string) *Common {
	ret := &Common{}
	if pos := strings.Index(actorFunc, "."); pos >= 0 {
		ret.actorName = actorFunc[:pos]
		ret.actorName = actorFunc[pos+1:]
	}
	return ret
}

func (d *Common) GetActorName() string {
	return d.actorName
}

func (d *Common) GetFuncName() string {
	return d.funcName
}

func (d *Common) AddDepth(val uint32) uint32 {
	return atomic.AddUint32(&d.depth, 1)
}

func (d *Common) CompareAndSwapDepth(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&d.depth, old, new)
}

func (d *Common) Router(idType int32, id uint64, routerId uint64) {
	d.idType = idType
	d.id = id
	d.routerId = routerId
}

func (d *Common) Rpc(nodeType int32, actorFunc string, actorId uint64) {
	d.dstNodeType = nodeType
	d.dstActorFunc = actorFunc
	d.dstActorId = actorId
}

func (d *Common) Callback(actorFunc string, actorId uint64) {
	d.cbActorFunc = actorFunc
	d.cbActorId = actorId
	d.cbNodeType = global.GetSelfType()
	d.cbNodeId = global.GetSelfId()
}
