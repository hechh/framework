package actor

import (
	"fmt"
	"github.com/hechh/framework/core/msgqueue"
	"github.com/hechh/framework/core/actor/internal/domain"
	"github.com/hechh/framework/core/actor/internal/task"
	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/packet"
	"reflect"
	"time"

	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
)

type ActorPool struct {
	domain.IAsync
	self domain.IActor
}

func (d *ActorPool) SetCache(string, any, uint32) {}
func (d *ActorPool) GetCache(string) (any, bool)  { return nil, false }
func (d *ActorPool) DelCache(string)              {}
func (d *ActorPool) IsChagnes(string) bool        { return false }
func (d *ActorPool) Reset(string)                 {}
func (d *ActorPool) Change(string)                {}
func (d *ActorPool) Init() error                  { return nil }
func (d *ActorPool) Close()                       {}

func (d *ActorPool) GetActor(actorId uint64) domain.IActor {
	return d.self
}

// Register 注册Actor
func (d *ActorPool) Register(ac domain.IActor, opts ...domain.Option) {
	p := &domain.Attribute{Name: utils.ParseName(reflect.TypeOf(ac))}
	for _, opt := range opts {
		opt(p)
	}
	if p.GetId() == 0 {
		p.SetId(domain.GenActorId())
	}
	if p.LockSecond > 0 {
		d.IAsync = async.NewAsyncPoolLocker(p)
	} else {
		d.IAsync = async.NewAsyncPool(p)
	}
	d.self = ac
}

// RegisterTimer 注册定时器
func (d *ActorPool) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.GetIdPointer(), ms, times, func() {
		head := packet.GetHead(fun.ACTOR(name, d.GetId()))
		if err := d.SendMsg(head); err != nil {
			mlog.Errorf("Actor定时器转发失败. error=%v", err)
		}
	})
	return timer.Register(task)
}

// SendMsg 发送消息
func (d *ActorPool) SendMsg(head *packet.Head, args ...any) error {
	handler := handler.Get(head)
	if handler == nil {
		return fmt.Errorf("Handler(%s)未注册", head.ActorFuncName)
	}
	if !d.Push(task.NewTask(d.self, handler, head, args)) {
		mlog.Errorf("ActorPool(%s)服务已经停止，请求丢失, head=%v, args=%v", handler.GetActorFuncName(), head, args)
		return fmt.Errorf("ActorPool(%s)已经停止服务", handler.GetActorFuncName())
	}
	return nil
}

// Send 发送RPC消息
func (d *ActorPool) Send(head *packet.Head, body []byte) error {
	api := rpc.GetByActorFunc(head.DstType, head.ActorFunc)
	if api == nil {
		return fmt.Errorf("ActorPool(%d)未注册", head.ActorFunc)
	}
	handler := handler.GetByActorFunc(head.ActorFunc)
	if handler == nil {
		return fmt.Errorf("ActorPool(%s)未注册", api.GetActorFuncName())
	}
	if !d.Push(task.NewRpcTask(d.self, handler, api, head, body)) {
		mlog.Errorf("ActorPool(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("ActorPool(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
