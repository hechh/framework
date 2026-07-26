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
	"sync"
	"time"

	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
)

type ActorGroup struct {
	domain.IAsync
	queue []domain.IAsync
	self  domain.IActor
}

func (d *ActorGroup) SetCache(string, any, uint32) {}
func (d *ActorGroup) GetCache(string) (any, bool)  { return nil, false }
func (d *ActorGroup) DelCache(string)              {}
func (d *ActorGroup) IsChagnes(string) bool        { return false }
func (d *ActorGroup) Reset(string)                 {}
func (d *ActorGroup) Change(string)                {}

func (d *ActorGroup) Init() error {
	for i, item := range d.queue {
		if !item.Start() {
			for j := 0; j < i; j++ {
				d.queue[j].Stop()
			}
			return fmt.Errorf("ActorGroup(%s)启动失败", d.GetName())
		}
	}
	return nil
}

func (d *ActorGroup) Close() {
	for _, item := range d.queue {
		item.Stop()
	}
	wg := sync.WaitGroup{}
	wg.Add(d.GetSize())
	for _, item := range d.queue {
		go func() {
			item.Wait()
			wg.Done()
		}()
	}
	wg.Wait()
}

func (d *ActorGroup) GetActor(actorId uint64) domain.IActor {
	return d.self
}

func (d *ActorGroup) Register(ac domain.IActor, opts ...domain.Option) {
	p := &domain.Attribute{Name: utils.ParseName(reflect.TypeOf(ac))}
	for _, opt := range opts {
		opt(p)
	}
	if p.GetId() == 0 {
		p.SetId(domain.GenActorId())
	}
	d.IAsync = async.NewAsyncPool(p)
	d.self = ac
	// 初始化队列池：每个子队列需要独立的 AsyncBase（Status/Id 等状态字段不可共享）
	d.queue = make([]domain.IAsync, p.Size)
	d.queue[0] = d.IAsync
	for i := 1; i < p.Size; i++ {
		item := &domain.Attribute{Name: p.Name}
		for _, opt := range opts {
			opt(item)
		}
		item.SetId(p.GetId())
		d.queue[i] = async.NewAsync(item)
	}
}

func (d *ActorGroup) RegisterTimer(name string, ms time.Duration, times int32) error {
	task := timer.NewTask(d.GetIdPointer(), ms, times, func() {
		head := packet.GetHead(fun.ACTOR(name, d.GetId()))
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

	actorId := domain.GetActorId(head)
	size := uint64(d.GetSize())
	if !d.queue[actorId%size].Push(task.NewTask(d.self, handler, head, args)) {
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
	actorId := domain.GetActorId(head)
	size := uint64(d.GetSize())
	if !d.queue[actorId%size].Push(task.NewRpcTask(d.self, handler, api, head, body)) {
		mlog.Errorf("ActorGroup(%s)服务已经停止，请求丢失, head=%v, body=%v", handler.GetActorFuncName(), head, body)
		return fmt.Errorf("ActorGroup(%s)服务已经停止", api.GetActorFuncName())
	}
	return nil
}
