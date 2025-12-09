package context

import (
	"fmt"
	"framework/library/mlog"
	"framework/library/util"
	"framework/packet"
	"strings"
	"sync/atomic"
)

type Default struct {
	*packet.Head
	actorName string // actor名字
	funcName  string // func名字
	depth     uint32 // 调用深度
}

func NewDefault(head *packet.Head, name string) *Default {
	pos := strings.Index(name, ".")
	actorName := ""
	if pos >= 0 {
		actorName = name[:pos]
	}
	return &Default{
		Head:      head,
		actorName: actorName,
		funcName:  name[pos+1:],
	}
}

func (d *Default) GetHead() *packet.Head {
	return d.Head
}

func (d *Default) GetActorId() uint64 {
	if d.ActorId <= 0 {
		return d.Id
	}
	return d.ActorId
}

func (d *Default) GetActorName() string {
	return d.actorName
}

func (d *Default) GetFuncName() string {
	return d.funcName
}

func (d *Default) AddDepth(val uint32) uint32 {
	return atomic.AddUint32(&d.depth, 1)
}

func (d *Default) CompareAndSwapDepth(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&d.depth, old, new)
}

func (d *Default) getformat(str string) string {
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

func (d *Default) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Default) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Default) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Default) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Default) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Default) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
