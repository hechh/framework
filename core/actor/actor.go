package actor

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hechh/framework/common/global"
	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/timer"
)

type IActor interface {
	GetName() string                                  // Actor名字
	GetId() uint64                                    // Actor ID
	SetId(uint64)                                     // Actor ID
	Stop()                                            // 关闭actor任务队列协程
	Wait()                                            // 等待actor协程退出
	Register(IActor, ...msgqueue.Option) bool         // 派生类自我注册
	RegisterTimer(string, time.Duration, int32) error // 注册定时器
	SendMsg(*packet.Head, ...any) error               // 异步调用派生类成员函数
	Send(*packet.Head, []byte) error                  // 异步调用派生类成员函数
}

type Actor struct {
	queue *msgqueue.MsgQueue[msgqueue.ITask]
	self  IActor
}

func (d *Actor) GetName() string  { return d.queue.GetName() }
func (d *Actor) GetId() uint64    { return d.queue.GetId() }
func (d *Actor) SetId(val uint64) { d.queue.SetId(val) }
func (d *Actor) Stop()            { d.queue.Stop() }
func (d *Actor) Wait()            { d.queue.Wait() }

func (d *Actor) Register(ac IActor, opts ...msgqueue.Option) bool {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, msgqueue.WithName(name))
	d.queue = msgqueue.NewMsgQueue[msgqueue.ITask]()
	if !d.queue.Start(opts...) {
		return false
	}
	d.self = ac
	return true
}

func (d *Actor) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.queue.GetIdPointer(), ms, times, func() {
		head := global.GetHead(fun.ACTOR(name, d.queue.GetId()))
		if err := d.SendMsg(head); err != nil {
			mlog.Errorf("Actor定时器转发失败. error=%v", err)
		}
	})
	return timer.Register(task)
}

func (d *Actor) SendMsg(head *packet.Head, args ...any) error {
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

func (d *Actor) Send(head *packet.Head, body []byte) error {
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
