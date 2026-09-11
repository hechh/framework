package actor

import (
	"runtime/debug"
	"sync"

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
		// recover 必须在 Release 之前：Release 会把 IContext 置 nil 并归还对象池，
		// 之后再访问 d 的方法/字段会二次 panic —— 真因被覆盖，且新 panic 会逃出 Do()
		// 终止队列协程（连带销毁整个 Actor，与"单任务崩溃不拖垮 Actor"的设计相悖）
		if e := recover(); e != nil {
			mlog.Fatalf("PANIC: %v\nStack Trace:\n%s", e, string(debug.Stack()))
			// panic 会跳过下方的 AutoRsp：CMD 请求补一次错误回包，避免客户端永久等待
			if logic.Has(d.GetMask(), define.CMD_FLAG) && d.GetDepth() == depth && len(d.args) > 0 {
				msgbus.AutoRsp(d.IContext, d.IHandler, d.ReadOnly(), d.args[len(d.args)-1], errTaskPanic)
			}
		}
		d.Release()
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
		d.Release()
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
