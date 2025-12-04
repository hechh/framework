package context

import (
	"fmt"
	"framework/define"
	"framework/library/mlog"
	"framework/packet"
	"sync/atomic"
)

type Remote struct {
	*packet.Head
	actorName string // actor名字
	funcName  string // func名字
	depth     uint32 // 调用深度
}

func (d *Remote) GetActorId() uint64 {
	if d.ActorId <= 0 {
		return d.Id
	}
	return d.ActorId
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

func (d *Remote) NewRpc() define.IRpc {
	return NewRpc(d.Head)
}

func (d *Remote) getformat(str string) string {
	if d.ActorId > 0 {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", d.SrcNodeType, d.SrcNodeId, d.DstNodeType, d.DstNodeId, d.Id, d.actorName, d.funcName, d.ActorId, str)
	} else {
		return fmt.Sprintf("[%d] Node(%d:%d) -> Node(%d:%d) %s.%s(%d)\t%s", d.Id, d.SrcNodeType, d.SrcNodeId, d.DstNodeType, d.DstNodeId, d.actorName, d.funcName, d.Id, str)
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
