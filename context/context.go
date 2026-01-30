package context

import (
	"strings"
	"sync/atomic"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/util"
)

type Context struct {
	*packet.Head
	actorFunc string // actor名字
	depth     int32  // 调用深度
}

func NewSimpleContext(aid uint64, name string) *Context {
	return &Context{
		Head: &packet.Head{
			Id:        aid,
			ActorFunc: framework.GetCrc32(name),
		},
		actorFunc: name,
	}
}

func NewContext(head *packet.Head, actorFunc string) *Context {
	return &Context{
		Head:      head,
		actorFunc: actorFunc,
	}
}

func (d *Context) To(actorFunc string) framework.IContext {
	d.ActorFunc = framework.GetCrc32(actorFunc)
	d.actorFunc = actorFunc
	return d
}

func (d *Context) Copy() framework.IContext {
	return &Context{
		Head: &packet.Head{
			SrcNodeType: framework.GetSelfType(),
			SrcNodeId:   framework.GetSelfId(),
			IdType:      d.Head.IdType,
			Id:          d.Head.Id,
			Version:     d.Version,
			SocketId:    d.SocketId,
			Extra:       d.Extra,
		},
	}
}

func (d *Context) GetHead() *packet.Head {
	return d.Head
}

func (d *Context) GetActorId() uint64 {
	return util.Or(d.ActorId > 0, d.ActorId, d.Id)
}

func (d *Context) GetActorName() string {
	return d.actorFunc[:strings.Index(d.actorFunc, ".")]
}

func (d *Context) GetFuncName() string {
	return d.actorFunc[strings.Index(d.actorFunc, ".")+1:]
}

func (d *Context) GetActorFunc() string {
	return d.actorFunc
}

func (d *Context) AddDepth(val int32) int32 {
	return atomic.AddInt32(&d.depth, val)
}

func (d *Context) CompareAndSwapDepth(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&d.depth, old, new)
}

func (d *Context) getformat(str string) string {
	return str
	/*
		return fmt.Sprintf("[%d], %s(%d), %s",
			d.Id,
			d.actorFunc,
			d.GetActorId(),
			str)
	*/
}

func (d *Context) Tracef(format string, args ...any) {
	mlog.Trace(1, d.GetFuncName(), d.getformat(format), args...)
}

func (d *Context) Debugf(format string, args ...any) {
	mlog.Debug(1, d.GetFuncName(), d.getformat(format), args...)
}

func (d *Context) Warnf(format string, args ...any) {
	mlog.Warn(1, d.GetFuncName(), d.getformat(format), args...)
}

func (d *Context) Infof(format string, args ...any) {
	mlog.Info(1, d.GetFuncName(), d.getformat(format), args...)
}

func (d *Context) Errorf(format string, args ...any) {
	mlog.Error(1, d.GetFuncName(), d.getformat(format), args...)
}

func (d *Context) Fatalf(format string, args ...any) {
	mlog.Fatal(1, d.GetFuncName(), d.getformat(format), args...)
}
