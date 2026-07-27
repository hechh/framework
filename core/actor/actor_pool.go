package actor

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/global"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/timer"
)

type ActorPool struct {
	queue *msgqueue.MsgQueuePool[msgqueue.ITask]
	self  IActor
}

func (d *ActorPool) GetName() string  { return d.queue.GetName() }
func (d *ActorPool) GetId() uint64    { return d.queue.GetId() }
func (d *ActorPool) SetId(val uint64) { d.queue.SetId(val) }
func (d *ActorPool) Stop()            { d.queue.Stop() }
func (d *ActorPool) Wait()            { d.queue.Wait() }

func (d *ActorPool) Register(ac IActor, opts ...msgqueue.Option) bool {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, msgqueue.WithName(name))
	d.queue = msgqueue.NewMsgQueuePool[msgqueue.ITask]()
	if !d.queue.Start(opts...) {
		return false
	}
	d.self = ac
	return true
}

func (d *ActorPool) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.queue.GetIdPointer(), ms, times, func() {
		head := global.GetHead(fun.ACTOR(name, d.queue.GetId()))
		if err := d.SendMsg(head); err != nil {
			mlog.Errorf("Actor定时器转发失败. error=%v", err)
		}
	})
	return timer.Register(task)
}

func (d *ActorPool) SendMsg(head *packet.Head, args ...any) error {
	handler := handler.Get(head)
	if handler == nil {
		return fmt.Errorf("Handler(%s)未注册", head.ActorFuncName)
	}
	if !d.queue.Push(NewTask(d.self, handler, head, args)) {
		mlog.Errorf("Actor(%s)服务已经停止，请求丢失, head=%v, args=%v", handler.GetActorFuncName(), head, args)
		return fmt.Errorf("Actor(%s)服务已经停止", handler.GetActorFuncName())
	}
	return nil
}

func (d *ActorPool) Send(head *packet.Head, body []byte) error {
	api := rpc.GetByActorFunc(head.DstType, head.ActorFunc)
	if api == nil {
		return fmt.Errorf("Rpc(%d)未注册", head.ActorFunc)
	}
	handler := handler.GetByActorFunc(head.ActorFunc)
	if handler == nil {
		return fmt.Errorf("Handler(%s)未注册", api.GetActorFuncName())
	}
	if !d.queue.Push(NewRpcTask(d.self, handler, api, head, body)) {
		mlog.Errorf("Actor(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("Actor(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
