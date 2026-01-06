package context

import (
	"fmt"
	"framework/core/define"
	"framework/library/mlog"
	"framework/library/util"
	"framework/packet"
	"strings"
	"sync/atomic"
)

type Context struct {
	head      *packet.Head // 包头
	actorName string       // actor名字
	funcName  string       // func名字
	depth     uint32       // 调用深度
}

func NewContext(head *packet.Head, actorName, funcName string) *Context {
	return &Context{
		head:      head,
		actorName: actorName,
		funcName:  funcName,
	}
}

func (d *Context) GetHead() *packet.Head {
	return d.head
}

func (d *Context) GetPacket() define.IPacket {
	d.AddDepth(1)
	return NewPacket(d.head)
}

func (d *Context) GetIdType() uint32 {
	return d.head.IdType
}

func (d *Context) GetId() uint64 {
	return d.head.Id
}

func (d *Context) GetActorId() uint64 {
	return util.Or(d.head.ActorId > 0, d.head.ActorId, d.head.Id)
}

func (d *Context) GetActorName() string {
	return d.actorName
}

func (d *Context) GetFuncName() string {
	return d.funcName
}

func (d *Context) IsRsp() bool {
	return d.head.Back != nil || d.head.Cmd > 0
}

func (d *Context) AddDepth(val uint32) uint32 {
	return atomic.AddUint32(&d.depth, 1)
}

func (d *Context) CompareAndSwapDepth(old, new uint32) bool {
	return atomic.CompareAndSwapUint32(&d.depth, old, new)
}

func (d *Context) To(actorFunc string) define.IContext {
	if pos := strings.Index(actorFunc, "."); pos > 0 {
		d.actorName = actorFunc[:pos]
		d.funcName = actorFunc[pos+1:]
	} else {
		d.actorName = ""
		d.funcName = actorFunc
	}
	return d
}

func (d *Context) getformat(str string) string {
	return fmt.Sprintf("[%d:%d->%d:%d] [%d] %s.%s(%d) %s",
		d.head.SrcNodeType,
		d.head.SrcNodeId,
		d.head.DstNodeType,
		d.head.DstNodeId,
		d.head.Id,
		d.actorName,
		d.funcName,
		util.Or(d.head.ActorId > 0, d.head.ActorId, d.head.Id),
		str)
}

func (d *Context) Tracef(format string, args ...any) {
	mlog.Trace(1, d.getformat(format), args...)
}

func (d *Context) Debugf(format string, args ...any) {
	mlog.Debug(1, d.getformat(format), args...)
}

func (d *Context) Warnf(format string, args ...any) {
	mlog.Warn(1, d.getformat(format), args...)
}

func (d *Context) Infof(format string, args ...any) {
	mlog.Info(1, d.getformat(format), args...)
}

func (d *Context) Errorf(format string, args ...any) {
	mlog.Error(1, d.getformat(format), args...)
}

func (d *Context) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.getformat(format), args...)
}
