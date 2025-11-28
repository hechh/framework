package context

import (
	"fmt"
	"framework/library/mlog"
	"framework/packet"
	"strings"
	"sync/atomic"
)

type Remote struct {
	*packet.Head
	actorName string
	funcName  string
	depth     uint32
	router    *Router
	callback  *Address
	dst       *Address
}

func NewRemote(head *packet.Head, actorFunc string) *Remote {
	ret := &Remote{Head: head}
	if pos := strings.Index(actorFunc, "."); pos >= 0 {
		ret.actorName = actorFunc[:pos]
		ret.actorName = actorFunc[pos+1:]
	}
	return ret
}

func (d *Remote) GetActorName() string {
	return d.actorName
}

func (d *Remote) GetFuncName() string {
	return d.funcName
}

func (d *Remote) AddDepth(val uint32) uint32 {
	return atomic.AddUint32(&d.depth, 1)
}

func (d *Remote) CompareAndSwapDepth(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&d.depth, old, new)
}

func (d *Remote) Router(idType int32, id uint64, routerId uint64) {
	d.router = &Router{idType, id, routerId}
}

func (d *Remote) Rpc(nodeType int32, actorFunc string, actorId uint64) {
	d.dst = &Address{nodeType, actorFunc, actorId}
}

func (d *Remote) Callback(actorFunc string, actorId uint64) {
	d.callback = &Address{
		actorFunc: actorFunc,
		actorId:   actorId,
	}
}

func (d *Remote) getformat(str string) string {
	if d.ActorId > 0 {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", d.Src.NodeType, d.Src.NodeId, d.Dst.NodeType, d.Dst.NodeId, d.Uid, d.actorName, d.funcName, d.ActorId, str)
	} else {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", d.Uid, d.Src.NodeType, d.Src.NodeId, d.Dst.NodeType, d.Dst.NodeId, d.actorName, d.funcName, d.Uid, str)
	}
}

func (d *Remote) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Remote) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Remote) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Remote) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Remote) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Remote) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
