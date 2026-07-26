package actor

import (
	"fmt"
	"github.com/hechh/framework/core/actor/internal/domain"
	"github.com/hechh/framework/packet"
	"reflect"
	"sync"
	"time"

	"github.com/hechh/library/base/utils"
)

type Shards struct {
	mutex  sync.RWMutex
	actors map[uint64]domain.IActor
}

type ActorMgr[T any] struct {
	*domain.Attribute
	shards []*Shards
}

func (d *ActorMgr[T]) SetCache(string, any, uint32) {}
func (d *ActorMgr[T]) GetCache(string) (any, bool)  { return nil, false }
func (d *ActorMgr[T]) DelCache(string)              {}
func (d *ActorMgr[T]) IsChagnes(string) bool        { return false }
func (d *ActorMgr[T]) Reset(string)                 {}
func (d *ActorMgr[T]) Change(string)                {}
func (d *ActorMgr[T]) Start() bool                  { return true }
func (d *ActorMgr[T]) Stop()                        {}
func (d *ActorMgr[T]) Wait()                        {}
func (d *ActorMgr[T]) Init() error                  { return nil }
func (d *ActorMgr[T]) Close()                       {}

func (d *ActorMgr[T]) GetActor(id uint64) domain.IActor {
	sh := d.shards[id%uint64(d.Size)]
	sh.mutex.RLock()
	act, ok := sh.actors[id]
	sh.mutex.RUnlock()
	if ok {
		return act
	}
	return nil
}

func (d *ActorMgr[T]) Register(ac domain.IActor, opts ...domain.Option) {
	p := &domain.Attribute{Name: utils.ParseName(reflect.TypeOf(ac))}
	for _, opt := range opts {
		opt(p)
	}
	if p.GetId() == 0 {
		p.SetId(domain.GenActorId())
	}
	d.Attribute = p
	d.shards = make([]*Shards, p.Size)
	for i := uint64(0); i < uint64(p.Size); i++ {
		d.shards[i] = &Shards{actors: make(map[uint64]domain.IActor)}
	}
}

func (d *ActorMgr[T]) RegisterTimer(name string, ms time.Duration, times int32) error {
	/*
		task := timer.NewTask(d.GetIdPointer(), ms, times, func() {
			head := packet.GetHead(fun.ACTOR(name, d.GetId()))
			if err := d.SendMsg(head); err != nil {
				mlog.Errorf("ActorMgr定时器转发失败. error=%v", err)
			}
		})
		return timer.Register(task)
	*/
	return nil
}

func (d *ActorMgr[T]) SendMsg(head *packet.Head, args ...any) error {
	switch head.SendType {
	case packet.SendType_POINT:
		id := domain.GetActorId(head)
		var err error
		act := d.GetActor(id)
		if act == nil {
			act, err = d.NewActor(id)
		}
		if err != nil {
			return err
		}
		return act.SendMsg(head, args...)
	case packet.SendType_BROADCAST:
		for _, sh := range d.shards {
			sh.mutex.RLock()
			for _, act := range sh.actors {
				act.SendMsg(head, args...)
			}
			sh.mutex.RUnlock()
		}
		return nil
	default:
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.Name, head.SendType)
	}
}

func (d *ActorMgr[T]) Send(head *packet.Head, body []byte) error {
	switch head.SendType {
	case packet.SendType_POINT:
		id := domain.GetActorId(head)
		var err error
		act := d.GetActor(id)
		if act == nil {
			act, err = d.NewActor(id)
		}
		if err != nil {
			return err
		}
		return act.Send(head, body)
	case packet.SendType_BROADCAST:
		for _, sh := range d.shards {
			sh.mutex.RLock()
			for _, act := range sh.actors {
				act.Send(head, body)
			}
			sh.mutex.RUnlock()
		}
		return nil
	default:
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.Name, head.SendType)
	}
}

func (d *ActorMgr[T]) DelActor(id uint64) {
	sh := d.shards[id%uint64(d.Size)]
	sh.mutex.Lock()
	delete(sh.actors, id)
	sh.mutex.Unlock()
}

func (d *ActorMgr[T]) NewActor(id uint64) (domain.IActor, error) {
	if act := d.GetActor(id); act != nil {
		return act, nil
	}

	sh := d.shards[id%uint64(d.Size)]
	sh.mutex.Lock()
	defer sh.mutex.RUnlock()

	// 二次获取
	if act, ok := sh.actors[id]; ok {
		return act, nil
	}

	// 初始化
	obj := any(new(T)).(domain.IActor)
	obj.SetId(id)
	if err := obj.Init(); err != nil {
		return nil, err
	}
	sh.actors[id] = obj
	return obj, nil
}
