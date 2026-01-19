package actor

import (
	"reflect"
	"sync/atomic"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/library/async"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/uerror"
	"github.com/hechh/library/util"
)

type ActorPool struct {
	self framework.IActor
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
	close(d.exit)
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

func (d *ActorPool) Register(ac framework.IActor, counts ...int) {
	d.id = framework.GenActorId()
	d.exit = make(chan struct{})
	d.size = util.Index[int](counts, 0, 10)
	d.pool = make([]*async.Async, d.size)
	for i := 0; i < d.size; i++ {
		d.pool[i] = async.NewAsync()
	}
	d.name = framework.ParseActorName(reflect.TypeOf(ac))
	d.self = ac
}

func (d *ActorPool) RegisterTimer(ctx framework.IContext, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(ctx); err != nil {
			ctx.Errorf("Actor定时器转发失败:%v", err)
		}
	})
}

func (d *ActorPool) SendMsg(ctx framework.IContext, args ...any) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.pool[ctx.GetActorId()%uint64(d.size)].Push(ff.Call(d.self, ctx, args...))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}

func (d *ActorPool) Send(ctx framework.IContext, body []byte) error {
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff != nil {
		ctx.AddDepth(1)
		d.pool[ctx.GetActorId()%uint64(d.size)].Push(ff.Rpc(d.self, ctx, body))
		return nil
	}
	return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
}
