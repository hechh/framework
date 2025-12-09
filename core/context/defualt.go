package context

import (
	"fmt"
	"framework/library/mlog"
	"framework/library/util"
	"framework/packet"
	"sync/atomic"
)

type Defualt struct {
	*packet.Head
	actorName string // actor名字
	funcName  string // func名字
	depth     uint32 // 调用深度
}

func (d *Defualt) GetActorId() uint64 {
	if d.ActorId <= 0 {
		return d.Id
	}
	return d.ActorId
}

func (d *Defualt) GetActorName() string {
	return d.actorName
}

func (d *Defualt) GetFuncName() string {
	return d.funcName
}

func (d *Defualt) AddDepth(val uint32) uint32 {
	return atomic.AddUint32(&d.depth, 1)
}

func (d *Defualt) CompareAndSwapDepth(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&d.depth, old, new)
}

func (d *Defualt) getformat(str string) string {
	return fmt.Sprintf("[%d:%d->%d:%d] [%d] %s.%s(%d) %s",
		d.SrcNodeType,
		d.SrcNodeId,
		d.DstNodeType,
		d.DstNodeId,
		d.Id,
		d.actorName,
		d.funcName,
		util.Or(d.ActorId > 0, d.ActorId, d.Id),
		str)
}

func (d *Defualt) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Defualt) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Defualt) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Defualt) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Defualt) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Defualt) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
