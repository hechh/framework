package actor

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hechh/framework/core/fun"
	"github.com/hechh/framework/core/handler"
	"github.com/hechh/framework/core/rpc"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/queue"
	"github.com/hechh/framework/library/utils"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/gc"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/timer"
)

type ActorPool struct {
	msgs  *queue.MsgQueuePool[queue.ITask]
	self  IActor
	cache define.ICache
}

func (d *ActorPool) GetName() string { return d.msgs.GetName() }
func (d *ActorPool) GetId() uint64   { return d.msgs.GetId() }
func (d *ActorPool) Start() error    { return d.msgs.Start(mlog.Fatalf) }

func (d *ActorPool) Stop() {
	d.msgs.Stop()
	gc.Destroy(d.msgs.Wait)
}

func (d *ActorPool) Register(ac IActor, c define.ICache, opts ...queue.Option) {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, queue.WithName(name))
	// opts 必须传给协程池：否则 size 默认 0 → 无缓冲队列且 0 个 worker，
	// 首个任务会永久阻塞在 taskCh，连带卡死 gc 单协程的全进程清理
	d.msgs = queue.NewMsgQueuePool[queue.ITask](opts...)
	if d.msgs.GetSize() <= 0 {
		panic(fmt.Sprintf("ActorPool(%s)协程池大小必须大于 0，请用 queue.WithSize 设置", name))
	}
	d.self = ac
	d.cache = c
}

func (d *ActorPool) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.msgs.GetIdPointer(), ms, times, func() {
		head := packet.GetHead(fun.ACTOR(name, d.GetId()))
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
	if !d.msgs.Push(NewTask(d.self, d.cache, handler, head, args)) {
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
	if !d.msgs.Push(NewRpcTask(d.self, d.cache, handler, api, head, body)) {
		mlog.Errorf("Actor(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("Actor(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
