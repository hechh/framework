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
	"github.com/hechh/library/base/cache"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/gc"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/timer"
)

type ActorGroup struct {
	head  *msgqueue.MsgQueue[msgqueue.ITask]
	msgs  []*msgqueue.MsgQueue[msgqueue.ITask]
	self  IActor
	cache cache.ICache
}

func (d *ActorGroup) GetName() string { return d.head.GetName() }
func (d *ActorGroup) GetId() uint64   { return d.head.GetId() }

func (d *ActorGroup) Start() bool {
	for i, act := range d.msgs {
		if !act.Start() {
			for j := i - 1; j >= 0; j-- {
				d.msgs[j].Stop()
				gc.Destroy(d.msgs[j].Wait)
			}
			return false
		}
	}
	return true
}

func (d *ActorGroup) Stop() {
	for _, item := range d.msgs {
		item.Stop()
		gc.Destroy(item.Wait)
	}
}

func (d *ActorGroup) Register(ac IActor, c cache.ICache, opts ...msgqueue.Option) {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, msgqueue.WithName(name))
	d.head = msgqueue.NewMsgQueue[msgqueue.ITask](opts...)
	d.self = ac
	d.cache = c
	d.msgs = make([]*msgqueue.MsgQueue[msgqueue.ITask], d.head.GetSize())
	d.msgs[0] = d.head
	for i := 1; i < d.head.GetSize(); i++ {
		d.msgs[i] = msgqueue.NewMsgQueue[msgqueue.ITask](opts...)
	}
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
	if !d.msgs[actorId%size].Push(NewTask(d.self, d.cache, handler, head, args)) {
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
	if !d.msgs[actorId%size].Push(NewRpcTask(d.self, d.cache, handler, api, head, body)) {
		mlog.Errorf("ActorGroup(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("ActorGroup(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
