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
)

type Actor struct {
	tasks *async.Async
	self  framework.IActor
	exit  chan struct{}
	name  string
}

func (d *Actor) Start() {
	d.tasks.Start()
}

func (d *Actor) Stop() {
	d.tasks.Stop()
}

func (d *Actor) Done() {
	d.tasks.Done()
}

func (d *Actor) Wait() {
	d.tasks.Wait()
}

func (d *Actor) GetActorName() string {
	return d.name
}

func (d *Actor) GetActorId() uint64 {
	return d.tasks.GetId()
}

func (d *Actor) SetActorId(id uint64) {
	d.tasks.SetId(id)
}

func (d *Actor) Register(ac framework.IActor, counts ...int) {
	d.name = framework.ParseActorName(reflect.TypeOf(ac))
	d.tasks = async.NewAsync()
	d.tasks.SetId(framework.GenActorId())
	d.exit = make(chan struct{})
	d.self = ac
}

func (d *Actor) RegisterTimer(name string, ms time.Duration, times int32) error {
	return timer.Register(d.tasks.GetIdPointer(), ms, times, func() {
		if err := d.SendMsg(context.NewSimpleContext(d.GetActorId(), name)); err != nil {
			mlog.Errorf("Actor定时器转发失败:%v", err)
		}
	})
}

func (d *Actor) SendMsg(ctx framework.IContext, args ...any) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Call(d.self, ctx, args...))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}

func (d *Actor) Send(ctx framework.IContext, body []byte) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Rpc(d.self, ctx, body))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}
