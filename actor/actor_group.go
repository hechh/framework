package actor

import (
	"reflect"
	"sync/atomic"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/uerror"
	"github.com/hechh/library/util"
)

type ActorGroup struct {
	id    uint64
	size  int
	queue []*async.Async
	self  framework.IActor
	exit  chan struct{}
	name  string
}

func (d *ActorGroup) Start() {
	for _, item := range d.queue {
		item.Start()
	}
}

func (d *ActorGroup) Stop() {
	for _, item := range d.queue {
		item.Stop()
	}
}

func (d *ActorGroup) Done() {
	for _, item := range d.queue {
		item.Done()
	}
}

func (d *ActorGroup) Wait() {
	for _, item := range d.queue {
		item.Wait()
	}
}

func (d *ActorGroup) GetActorName() string {
	return d.name
}

func (d *ActorGroup) GetActorId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *ActorGroup) SetActorId(id uint64) {
	atomic.StoreUint64(&d.id, id)
}

func (d *ActorGroup) Register(ac framework.IActor, counts ...int) {
	d.name = framework.ParseActorName(reflect.TypeOf(ac))
	d.size = util.Index(counts, 0, 10)
	d.queue = make([]*async.Async, d.size)
	for i := range d.queue {
		d.queue[i] = async.NewAsync()
	}
	d.SetActorId(framework.GenActorId())
	d.exit = make(chan struct{})
	d.self = ac
}

func (d *ActorGroup) RegisterTimer(name string, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(context.NewSimpleContext(d.GetActorId(), name)); err != nil {
			mlog.Errorf("ActorGroup定时器转发失败:%v", err)
		}
	})
}

func (d *ActorGroup) SendMsg(ctx framework.IContext, args ...any) error {
	var err error
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff == nil {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	} else {
		if !d.queue[ctx.GetActorId()%uint64(d.size)].Push(ff.Call(d.self, ctx, args...)) {
			err = uerror.Err(-1, "ActorGroup已经停止服务")
		}
	}
	ctx.Tracef("[actor] ActorGroup(%s)本地调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, args)
	return err
}

func (d *ActorGroup) Send(ctx framework.IContext, body []byte) error {
	var err error
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff == nil {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	} else {
		if !d.queue[ctx.GetActorId()%uint64(d.size)].Push(ff.Rpc(d.self, ctx, body)) {
			err = uerror.Err(-1, "ActorGroup已经停止服务")
		}
	}
	ctx.Tracef("[actor] ActorGroup(%s)远程调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, body)
	return err
}
