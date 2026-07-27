package actor

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hechh/framework/common/global"
	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/timer"
)

type ActorGroup struct {
	head   *msgqueue.MsgQueue[msgqueue.ITask]
	queues []*msgqueue.MsgQueue[msgqueue.ITask]
	self   IActor
}

func (d *ActorGroup) GetName() string  { return d.queues[0].GetName() }
func (d *ActorGroup) GetId() uint64    { return d.queues[0].GetId() }
func (d *ActorGroup) SetId(val uint64) { d.queues[0].SetId(val) }

func (d *ActorGroup) Stop() {
	for _, item := range d.queues {
		item.Stop()
	}
}

func (d *ActorGroup) Wait() {
	wg := sync.WaitGroup{}
	wg.Add(d.head.GetSize())
	for _, item := range d.queues {
		go func() {
			item.Wait()
			wg.Done()
		}()
	}
	wg.Wait()
	d.queues[0].Wait()
}

func (d *ActorGroup) Register(ac IActor, opts ...msgqueue.Option) bool {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, msgqueue.WithName(name), msgqueue.WithId(msgqueue.GenId()))

	d.head = msgqueue.NewMsgQueue[msgqueue.ITask]()
	if !d.head.Start(opts...) {
		return false
	}
	d.self = ac

	d.queues = make([]*msgqueue.MsgQueue[msgqueue.ITask], d.head.GetSize())
	d.queues[0] = d.head
	for i := 1; i < d.head.GetSize(); i++ {
		d.queues[i] = msgqueue.NewMsgQueue[msgqueue.ITask]()
		if !d.queues[i].Start(opts...) {
			// 失败时，关闭
			for j := i - 1; j >= 0; j-- {
				d.queues[j].Stop()
			}
			return false
		}
	}
	return true
}

func (d *ActorGroup) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.head.GetIdPointer(), ms, times, func() {
		head := global.GetHead(fun.ACTOR(name, d.GetId()))
		if err := d.SendMsg(head); err != nil {
			mlog.Errorf("Actor定时器转发失败. error=%v", err)
		}
	})
	return timer.Register(task)
}

func (d *ActorGroup) SendMsg(head *packet.Head, args ...any) error {
	handler := handler.Get(head)
	if handler == nil {
		return fmt.Errorf("ActorGroup(%s)未注册", head.ActorFuncName)
	}
	actorId := templ.Or(head.ActorId > 0, head.ActorId, head.Uid)
	size := uint64(d.head.GetSize())
	if !d.queues[actorId%size].Push(NewTask(d.self, handler, head, args)) {
		mlog.Errorf("ActorGroup(%s)服务已经停止，请求丢失, head=%v, args=%v", handler.GetActorFuncName(), head, args)
		return fmt.Errorf("ActorGroup(%s)服务已经停止", handler.GetActorFuncName())
	}
	return nil
}

func (d *ActorGroup) Send(head *packet.Head, body []byte) error {
	api := rpc.GetByActorFunc(head.DstType, head.ActorFunc)
	if api == nil {
		return fmt.Errorf("ActorGroup(%d)未注册", head.ActorFunc)
	}
	handler := handler.GetByActorFunc(head.ActorFunc)
	if handler == nil {
		return fmt.Errorf("ActorGroup(%s)未注册", api.GetActorFuncName())
	}
	actorId := templ.Or(head.ActorId > 0, head.ActorId, head.Uid)
	size := uint64(d.head.GetSize())
	if !d.queues[actorId%size].Push(NewRpcTask(d.self, handler, api, head, body)) {
		mlog.Errorf("ActorGroup(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("ActorGroup(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
