package context

import (
	"fmt"
	"sync/atomic"

	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Context struct {
	*packet.Head
	temps map[string]define.IValue
	cache define.ICache
}

func NewContext(val any, data define.ICache, opts ...func(*packet.Head)) *Context {
	var head *packet.Head
	switch vv := val.(type) {
	case *packet.Head:
		head = vv
	case uint64:
		head = packet.GetHead(fun.UID(vv))
	}
	for _, opt := range opts {
		opt(head)
	}
	return &Context{
		Head:  head,
		temps: make(map[string]define.IValue),
		cache: data,
	}
}

func (c *Context) ReadOnly() *packet.Head {
	return c.Head
}

func (c *Context) Clone(opts ...func(*packet.Head)) *packet.Head {
	head := packet.GetHead(fun.COPY(c.Head))
	for _, opt := range opts {
		opt(head)
	}
	return head
}

func (c *Context) Derive(opts ...func(*packet.Head)) *packet.Head {
	head := packet.GetHead(fun.DERIVE(c.Head))
	for _, opt := range opts {
		opt(head)
	}
	return head
}

func (c *Context) Has(key string) bool {
	if _, ok := c.temps[key]; ok {
		return ok
	}
	return c.cache.Has(key)
}

func (c *Context) SetCache(key string, value define.IValue) {
	c.temps[key] = value
}

func (c *Context) GetCache(key string) define.IValue {
	if val, ok := c.temps[key]; ok {
		return val
	}
	// 常驻缓存，GetCache需要深度拷贝
	if c.cache.Has(key) {
		vv := c.cache.GetCache(key).Clone()
		c.temps[key] = vv
		return vv
	}
	return nil
}

func (c *Context) GetAllCache() map[string]define.IValue {
	return c.temps
}

// Refresh 把临时缓存中已变更的值提交到常驻缓存；except 中的 key 不提交。
// 部分写入失败（如某个 Redis 分组失败）时要把失败数据排除在外：常驻缓存必须与
// Redis 保持一致，否则下次 Save 会用未持久化的新值覆盖已写成功的分组。
func (c *Context) Refresh(except ...string) {
	var skip map[string]struct{}
	if len(except) > 0 {
		skip = make(map[string]struct{}, len(except))
		for _, k := range except {
			skip[k] = struct{}{}
		}
	}
	for k, v := range c.temps {
		if !v.IsChanged() {
			continue
		}
		if _, ok := skip[k]; ok {
			continue
		}
		if c.cache.Has(k) {
			c.cache.SetCache(k, v)
		}
		v.Reset()
	}
}

func (c *Context) AddDepth(val int32) int32      { return atomic.AddInt32(&c.Depth, val) }
func (c *Context) GetDepth() int32               { return atomic.LoadInt32(&c.Depth) }
func (c *Context) GetActorId() uint64            { return tplutil.Or(c.ActorId > 0, c.ActorId, c.Uid) }
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
	return fmt.Sprintf("%dms|%d", now-datetime.NanoToMilli(head.CreateTime), head.TraceId)
}
