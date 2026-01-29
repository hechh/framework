package actor

import (
	"reflect"
	"strings"
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
	var err error
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff == nil {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	} else {
		if !d.tasks.Push(ff.Call(d.self, ctx, args...)) {
			err = uerror.Err(-1, "Actor已经停止服务")
		}
	}
	if !strings.HasSuffix(ctx.GetActorFunc(), "OnTick") {
		mlog.Trace(-1, "[actor] Actor(%s)本地调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, args)
	}
	return err
}

func (d *Actor) Send(ctx framework.IContext, body []byte) error {
	var err error
	if ff := framework.GetHandler(ctx.GetActorFunc()); ff == nil {
		err = uerror.Err(-1, "%s未注册", ctx.GetActorFunc())
	} else {
		if !d.tasks.Push(ff.Rpc(d.self, ctx, body)) {
			err = uerror.Err(-1, "Actor已经停止服务")
		}
	}
	if !strings.HasSuffix(ctx.GetActorFunc(), "OnTick") {
		mlog.Trace(-1, "[actor] Actor(%s)远程调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, body)
	}
	return err
}
