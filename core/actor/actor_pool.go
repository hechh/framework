package actor

import (
	"framework/core"
	"framework/library/async"
	"framework/library/timer"
	"framework/library/uerror"
	"framework/library/util"
	"reflect"
	"sync/atomic"
	"time"
)

type ActorPool struct {
	self core.IActor
	pool []*async.Async
	exit chan struct{}
	size int
	id   uint64
	name string
}

func (d *ActorPool) Start() {
	for _, act := range d.pool {
		act.Start()
	}
}

func (d *ActorPool) Stop() {
	atomic.StoreUint64(&d.id, 0)
	for _, act := range d.pool {
		act.Stop()
	}
}

func (d *ActorPool) Done() {
	select {
	case d.exit <- struct{}{}:
	default:
	}
}

func (d *ActorPool) Wait() {
	<-d.exit
}

func (d *ActorPool) GetActorName() string {
	return d.name
}

func (d *ActorPool) GetActorId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *ActorPool) SetActorId(id uint64) {
	atomic.StoreUint64(&d.id, id)
}

func (d *ActorPool) Register(ac core.IActor, counts ...int) {
	d.id = core.GenerateActorId()
	d.exit = make(chan struct{}, 1)
	d.size = util.Index[int](counts, 0, 10)
	d.pool = make([]*async.Async, d.size)
	for i := 0; i < d.size; i++ {
		d.pool[i] = async.NewAsync()
	}
	d.name = core.ParseActorName(reflect.TypeOf(ac))
	d.self = ac
}

func (d *ActorPool) RegisterTimer(ctx core.IContext, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(ctx); err != nil {
			ctx.Errorf("Actor定时器转发失败:%v", err)
		}
	})
}

func (d *ActorPool) SendMsg(ctx core.IContext, args ...any) error {
	if ff := core.GetHandler(ctx.GetActorName(), ctx.GetFuncName()); ff != nil {
		ctx.AddDepth(1)
		d.pool[ctx.GetActorId()%uint64(d.size)].Push(ff.Call(d.self, ctx, args...))
		return nil
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}

func (d *ActorPool) Send(ctx core.IContext, body []byte) error {
	if ff := core.GetHandler(ctx.GetActorName(), ctx.GetFuncName()); ff != nil {
		ctx.AddDepth(1)
		d.pool[ctx.GetActorId()%uint64(d.size)].Push(ff.Rpc(d.self, ctx, body))
		return nil
	}
	return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
}
