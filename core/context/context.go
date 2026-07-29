package context

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/global"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/mlog"
)

var (
	ctxPool = sync.Pool{
		New: func() any {
			return new(Context)
		},
	}
)

type Value struct {
	data  any
	times uint32
}

type Context struct {
	*packet.Head
	values map[string]*Value
}

func NewContext(head *packet.Head, opts ...func(*packet.Head)) *Context {
	for _, opt := range opts {
		opt(head)
	}
	obj := ctxPool.Get().(*Context)
	obj.Head = head
	obj.values = make(map[string]*Value)
	return obj
}

func (c *Context) Destroy() {
	if c.Head != nil {
		global.PutHead(c.Head)
		c.Head = nil
	}
	c.values = nil
	ctxPool.Put(c)
}

func (c *Context) ReadOnly() *packet.Head {
	return c.Head
}

func (c *Context) Clone(opts ...func(*packet.Head)) *packet.Head {
	head := global.GetHead(fun.COPY(c.Head))
	for _, opt := range opts {
		opt(head)
	}
	return head
}

func (c *Context) Derive(opts ...func(*packet.Head)) *packet.Head {
	head := global.GetHead(fun.DERIVE(c.Head))
	for _, opt := range opts {
		opt(head)
	}
	return head
}

func (c *Context) SetCache(key string, value any) {
	if v, ok := c.values[key]; ok {
		v.data = value
		return
	}
	c.values[key] = &Value{data: value}
}

func (c *Context) GetCache(key string) (any, bool) {
	if v, ok := c.values[key]; ok {
		return v.data, ok
	}
	return nil, false
}

func (c *Context) IsChanged(key string) bool {
	if v, ok := c.values[key]; ok {
		return v.times > 0
	}
	return false
}

func (c *Context) Change(key string) {
	if v, ok := c.values[key]; ok {
		v.times++
	}
}

func (c *Context) Reset(key string) {
	if v, ok := c.values[key]; ok {
		v.times = 0
		return
	}
}

func (c *Context) AddDepth(val int32) int32      { return atomic.AddInt32(&c.Depth, val) }
func (c *Context) GetDepth() int32               { return atomic.LoadInt32(&c.Depth) }
func (c *Context) GetActorId() uint64            { return templ.Or(c.ActorId > 0, c.ActorId, c.Uid) }
func (c *Context) SetSendType(v packet.SendType) { c.Head.SendType = v }
func (c *Context) SetSrcType(v uint32)           { c.Head.SrcType = v }
func (c *Context) SetSrcId(v uint32)             { c.Head.SrcId = v }
func (c *Context) SetDstType(v uint32)           { c.Head.DstType = v }
func (c *Context) SetDstId(v uint32)             { c.Head.DstId = v }
func (c *Context) SetUid(v uint64)               { c.Head.Uid = v }
func (c *Context) SetCmd(v uint32)               { c.Head.Cmd = v }
func (c *Context) SetSeq(v uint32)               { c.Head.Seq = v }
func (c *Context) SetActorFunc(v uint32)         { c.Head.ActorFunc = v }
func (c *Context) SetActorId(v uint64)           { c.Head.ActorId = v }
func (c *Context) SetSocketId(v uint32)          { c.Head.SocketId = v }
func (c *Context) SetVersion(v uint32)           { c.Head.Version = v }
func (c *Context) SetTraceId(v uint32)           { c.Head.TraceId = v }
func (c *Context) SetCreateTime(v int64)         { c.Head.CreateTime = v }
func (c *Context) SetBack(v *packet.Callback)    { c.Head.Back = v }
func (c *Context) SetReply(v string)             { c.Head.Reply = v }
func (c *Context) SetActorFuncName(v string)     { c.Head.ActorFuncName = v }
func (c *Context) GetClientIp() string           { return c.Head.ClientIp }
func (c *Context) SetClientIp(ip string)         { c.Head.ClientIp = ip }

func (c *Context) Trace(args ...any) { mlog.Output(2, mlog.LOG_TRACE, tag(c.Head), args...) }
func (c *Context) Debug(args ...any) { mlog.Output(2, mlog.LOG_DEBUG, tag(c.Head), args...) }
func (c *Context) Warn(args ...any)  { mlog.Output(2, mlog.LOG_WARN, tag(c.Head), args...) }
func (c *Context) Info(args ...any)  { mlog.Output(2, mlog.LOG_INFO, tag(c.Head), args...) }
func (c *Context) Error(args ...any) { mlog.Output(2, mlog.LOG_ERROR, tag(c.Head), args...) }
func (c *Context) Fatal(args ...any) { mlog.Output(2, mlog.LOG_FATAL, tag(c.Head), args...) }

func (c *Context) Tracef(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_TRACE, tag(c.Head), f, args...)
}
func (c *Context) Debugf(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_DEBUG, tag(c.Head), f, args...)
}
func (c *Context) Warnf(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_WARN, tag(c.Head), f, args...)
}
func (c *Context) Infof(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_INFO, tag(c.Head), f, args...)
}
func (c *Context) Errorf(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_ERROR, tag(c.Head), f, args...)
}
func (c *Context) Fatalf(f string, args ...any) {
	mlog.Outputf(2, mlog.LOG_FATAL, tag(c.Head), f, args...)
}

func tag(head *packet.Head) string {
	now := datetime.NowUnixMilli()
	return fmt.Sprintf("%dms|%s", now-datetime.NanoToMilli(head.CreateTime), head.TraceId)
}
