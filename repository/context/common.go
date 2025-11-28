package context

import (
	"fmt"
	"framework/library/mlog"
	"strings"
	"sync/atomic"
)

type Router struct {
	idType   int32
	id       uint64
	routerId uint64
}

type Address struct {
	nodeType  int32
	actorFunc string
	actorId   uint64
}

type Common struct {
	uid       uint64
	actorId   uint64
	actorName string
	funcName  string
	depth     uint32
	router    *Router
	callback  *Address
	dst       *Address
}

func NewCommon(uid uint64, actorFunc string, actorId uint64) *Common {
	ret := &Common{uid: uid, actorId: actorId}
	if pos := strings.Index(actorFunc, "."); pos >= 0 {
		ret.actorName = actorFunc[:pos]
		ret.actorName = actorFunc[pos+1:]
	}
	return ret
}

func (d *Common) GetUid() uint64 {
	return d.uid
}

func (d *Common) GetActorId() uint64 {
	return d.actorId
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
	d.router = &Router{idType, id, routerId}
}

func (d *Common) Rpc(nodeType int32, actorFunc string, actorId uint64) {
	d.dst = &Address{nodeType, actorFunc, actorId}
}

func (d *Common) Callback(actorFunc string, actorId uint64) {
	d.callback = &Address{
		actorFunc: actorFunc,
		actorId:   actorId,
	}
}

func (d *Common) getformat(str string) string {
	if d.uid > 0 {
		if d.actorId > 0 {
			return fmt.Sprintf("[%d] %s.%s(%d)\t%s", d.uid, d.actorName, d.funcName, d.actorId, str)
		} else {
			return fmt.Sprintf("[%d] %s.%s(%d)\t%s", d.uid, d.actorName, d.funcName, d.uid, str)
		}
	} else if d.actorId > 0 {
		return fmt.Sprintf("%s.%s(%d)\t%s", d.actorName, d.funcName, d.actorId, str)
	}
	return fmt.Sprintf("%s.%s\t%s", d.actorName, d.funcName, str)
}

func (d *Common) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Common) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Common) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Common) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Common) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Common) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
