package actor

import (
	"runtime/debug"

	"github.com/hechh/framework/core/context"
	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/logic"
	"github.com/hechh/framework/library/uerror"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

// errTaskPanic Task 执行 panic 时回给客户端的错误（框架层无业务错误码，-1 与裸 error 一致）
var errTaskPanic = uerror.Err(-1, "服务器内部错误")

type Task struct {
	define.IContext
	handler.IHandler
	actor any
	args  []any
}

func NewTask(a any, c define.ICache, h handler.IHandler, head *packet.Head, args []any) *Task {
	return &Task{
		IContext: context.NewContext(head, c, fun.TRACE),
		IHandler: h,
		actor:    a,
		args:     args,
	}
}

func (d *Task) Do() (flag bool) {
	depth := d.AddDepth(1)
	flag = !logic.Has(d.GetMask(), define.UPDATETIME_MASK)
	defer func() {
		if e := recover(); e != nil {
			mlog.Fatalf("PANIC: %v\nStack Trace:\n%s", e, string(debug.Stack()))
			if logic.Has(d.GetMask(), define.CMD_FLAG) && d.GetDepth() == depth && len(d.args) > 0 {
				msgbus.AutoRsp(d.IContext, d.IHandler, d.ReadOnly(), d.args[len(d.args)-1], errTaskPanic)
			}
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
	return &RpcTask{
		IContext: context.NewContext(head, c, fun.TRACE),
		IHandler: h,
		r:        r,
		actor:    a,
		body:     body,
	}
}

func (d *RpcTask) Do() (flag bool) {
	depth := d.AddDepth(1)
	flag = !logic.Has(d.GetMask(), define.UPDATETIME_MASK)

	// 提前声明：panic 时 defer 需要用它补错误回包（panic 可能发生在 News 之前/之中）
	var (
		args []any
		err  error
	)
	defer func() {
		// 同 Task.Do：recover 必须在 Release 之前，否则二次 panic 会逃出 Do()
		if e := recover(); e != nil {
			mlog.Fatalf("PANIC: %v\nStack Trace:\n%s", e, string(debug.Stack()))
			// panic 会跳过下方的 AutoRsp：CMD 请求补一次错误回包，避免客户端永久等待
			if logic.Has(d.GetMask(), define.CMD_FLAG) && d.GetDepth() == depth && len(args) > 0 {
				msgbus.AutoRsp(d.IContext, d.IHandler, d.ReadOnly(), args[len(args)-1], errTaskPanic)
			}
		}
	}()

	// 解析参数
	mask := d.GetMask()
	args, err = d.r.News(d.body)
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
