package actor

import (
	"reflect"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/uerror"
	"github.com/hechh/library/util"
)

type ActorPool struct {
	tasks *async.AsyncPool
	self  framework.IActor
	exit  chan struct{}
	name  string
}

func (d *ActorPool) Start() {
	d.tasks.Start()
}

func (d *ActorPool) Stop() {
	d.tasks.Stop()
}

func (d *ActorPool) Done() {
	d.tasks.Done()
}

func (d *ActorPool) Wait() {
	d.tasks.Wait()
}

func (d *ActorPool) GetActorName() string {
	return d.name
}

func (d *ActorPool) GetActorId() uint64 {
	return d.tasks.GetId()
}

func (d *ActorPool) SetActorId(id uint64) {
	d.tasks.SetId(id)
}

func (d *ActorPool) Register(ac framework.IActor, counts ...int) {
	d.name = framework.ParseActorName(reflect.TypeOf(ac))
	d.tasks = async.NewAsyncPool(util.Index(counts, 0, 10))
	d.tasks.SetId(framework.GenActorId())
	d.exit = make(chan struct{})
	d.self = ac
}

func (d *ActorPool) RegisterTimer(name string, ms time.Duration, times int32) error {
	return timer.Register(d.tasks.GetIdPointer(), ms, times, func() {
		if err := d.SendMsg(context.NewSimpleContext(d.GetActorId(), name)); err != nil {
			mlog.Errorf("Actor定时器转发失败:%v", err)
		}
	})
}

func (d *ActorPool) SendMsg(ctx framework.IContext, args ...any) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Call(d.self, ctx, args...))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}

func (d *ActorPool) Send(ctx framework.IContext, body []byte) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Rpc(d.self, ctx, body))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}
