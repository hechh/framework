package actor

import (
	"framework/core/define"
	"framework/core/global"
	"framework/core/handler"
	"framework/library/async"
	"framework/library/timer"
	"framework/library/uerror"
	"reflect"
	"time"
)

type Actor struct {
	tasks *async.Async
	self  define.IActor
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
	select {
	case d.exit <- struct{}{}:
	default:
	}
}

func (d *Actor) Wait() {
	<-d.exit
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

func (d *Actor) Register(ac define.IActor, counts ...int) {
	d.name = global.ParseActorName(reflect.TypeOf(ac))
	d.tasks = async.NewAsync()
	d.tasks.SetId(global.GenerateActorId())
	d.exit = make(chan struct{}, 1)
	d.self = ac
}

func (d *Actor) RegisterTimer(ctx define.IContext, ms time.Duration, times int32) error {
	return timer.Register(d.tasks.GetIdPointer(), ms, times, func() {
		if err := d.SendMsg(ctx); err != nil {
			ctx.Errorf("Actor定时器转发失败:%v", err)
		}
	})
}

func (d *Actor) SendMsg(ctx define.IContext, args ...any) error {
	if ff := handler.Get(ctx.GetActorName(), ctx.GetFuncName()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Call(d.self, ctx, args...))
		return nil
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

func (d *Actor) Send(ctx define.IContext, body []byte) error {
	if ff := handler.Get(ctx.GetActorName(), ctx.GetFuncName()); ff != nil {
		ctx.AddDepth(1)
		d.tasks.Push(ff.Rpc(d.self, ctx, body))
		return nil
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}
