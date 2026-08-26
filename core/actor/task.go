package actor

import (
	"runtime/debug"
	"sync"

	"github.com/hechh/framework/core/context"
	"github.com/hechh/framework/core/define"
	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/library/logic"
	"github.com/hechh/framework/packet"
)

var (
	localPool = &sync.Pool{
		New: func() any { return new(Task) },
	}
	rpcPool = &sync.Pool{
		New: func() any {
			return new(RpcTask)
		},
	}
)

type Task struct {
	define.IContext
	handler.IHandler
	actor any
	args  []any
}

func NewTask(a any, c define.ICache, h handler.IHandler, head *packet.Head, args []any) *Task {
	t := localPool.Get().(*Task)
	t.IContext = context.NewContext(head, c, fun.TRACE)
	t.IHandler = h
	t.args = args
	t.actor = a
	return t
}

func (d *Task) Release() {
	d.IContext.Destroy()
	d.IContext = nil
	d.IHandler = nil
	d.args = nil
	d.actor = nil
	localPool.Put(d)
}

func (d *Task) Do() (flag bool) {
	depth := d.AddDepth(1)
	flag = !logic.Has(d.GetMask(), define.UPDATETIME_MASK)
	defer func() {
		d.Release()
		if err := recover(); err != nil {
			d.Fatalf("PANIC: %v\nStack Trace:\n%s", err, string(debug.Stack()))
		}
	}()

	mask := d.GetMask()
	err := d.Call(d.actor, d.IContext, d.args...)
	if err != nil {
		d.Error(err, d.args)
	} else if !logic.Has(mask, define.LOG_MASK) {
		d.Trace(d.args...)
	}

	// 是否自动回复
	if logic.Has(mask, define.CMD_FLAG) && d.GetDepth() == depth {
		msgbus.AutoRsp(d.IContext, d.IHandler, d.ReadOnly(), d.args[len(d.args)-1], err)
	}
	return
}

type RpcTask struct {
	define.IContext
	handler.IHandler
	r     rpc.IRpc
	actor any
	body  []byte
}

func NewRpcTask(a any, c define.ICache, h handler.IHandler, r rpc.IRpc, head *packet.Head, body []byte) *RpcTask {
	t := rpcPool.Get().(*RpcTask)
	t.IContext = context.NewContext(head, c, fun.TRACE)
	t.IHandler = h
	t.r = r
	t.actor = a
	t.body = body
	return t
}

func (d *RpcTask) Release() {
	d.IContext.Destroy()
	d.IContext = nil
	d.IHandler = nil
	d.r = nil
	d.actor = nil
	d.body = nil
	rpcPool.Put(d)
}

func (d *RpcTask) Do() (flag bool) {
	depth := d.AddDepth(1)
	flag = !logic.Has(d.GetMask(), define.UPDATETIME_MASK)
	defer func() {
		d.Release()
		if err := recover(); err != nil {
			d.Fatalf("PANIC: %v\nStack Trace:\n%s", err, string(debug.Stack()))
		}
	}()

	// 解析参数
	mask := d.GetMask()
	args, err := d.r.News(d.body)
	if err != nil {
		d.Error(err, args)
		return
	}

	err = d.Call(d.actor, d.IContext, args...)
	if err != nil {
		d.Error(err, args)
	} else if !logic.Has(mask, define.LOG_MASK) {
		d.Trace(args...)
	}

	// 是否自动回复
	if logic.Has(mask, define.CMD_FLAG) && d.GetDepth() == depth {
		msgbus.AutoRsp(d.IContext, d.IHandler, d.ReadOnly(), args[len(args)-1], err)
	}
	return
}
