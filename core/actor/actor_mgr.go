package actor

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
)

type Shards struct {
	mutex  sync.RWMutex
	actors map[uint64]IActor
}

type ActorMgr[T any] struct {
	attr   *msgqueue.Attribute
	shards []*Shards
}

func (d *ActorMgr[T]) GetName() string  { return d.attr.GetName() }
func (d *ActorMgr[T]) GetId() uint64    { return d.attr.GetId() }
func (d *ActorMgr[T]) SetId(val uint64) { d.attr.SetId(val) }
func (d *ActorMgr[T]) Stop()            {}
func (d *ActorMgr[T]) Wait()            {}

func (d *ActorMgr[T]) Register(ac IActor, opts ...msgqueue.Option) {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, msgqueue.WithName(name))
	d.attr = new(msgqueue.Attribute)
	for _, opt := range opts {
		opt(d.attr)
	}
	d.shards = make([]*Shards, d.attr.GetSize())
	for i := uint64(0); i < uint64(d.attr.GetSize()); i++ {
		d.shards[i] = &Shards{actors: make(map[uint64]IActor)}
	}
}

func (d *ActorMgr[T]) RegisterTimer(name string, ms time.Duration, times int32) error {
	return nil
}

func (d *ActorMgr[T]) SendMsg(head *packet.Head, args ...any) error {
	switch head.SendType {
	case packet.SendType_POINT:
		id := templ.Or(head.ActorId > 0, head.ActorId, head.Uid)
		var err error
		act := d.GetActor(id)
		if act == nil {
			act, err = d.NewActor(id)
		}
		if err == nil {
			err = act.SendMsg(head, args...)
		}
		if err != nil {
			mlog.Errorf("ActorId(%d)消息转发失败 error=%v", id, err)
		}
		return err
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
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.attr.GetName(), head.SendType)
	}
}

func (d *ActorMgr[T]) Send(head *packet.Head, body []byte) error {
	switch head.SendType {
	case packet.SendType_POINT:
		id := templ.Or(head.ActorId > 0, head.ActorId, head.Uid)
		var err error
		act := d.GetActor(id)
		if act == nil {
			act, err = d.NewActor(id)
		}
		if err == nil {
			err = act.Send(head, body)
		}
		if err != nil {
			mlog.Errorf("ActorId(%d)消息转发失败 error=%v", id, err)
		}
		return err
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
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.attr.GetName(), head.SendType)
	}
}

func (d *ActorMgr[T]) GetActor(id uint64) IActor {
	sh := d.shards[id%uint64(d.attr.GetSize())]
	sh.mutex.RLock()
	act, ok := sh.actors[id]
	sh.mutex.RUnlock()
	if ok {
		return act
	}
	return nil
}

func (d *ActorMgr[T]) DelActor(id uint64) {
	sh := d.shards[id%uint64(d.attr.GetSize())]
	sh.mutex.Lock()
	delete(sh.actors, id)
	sh.mutex.Unlock()
}

func (d *ActorMgr[T]) NewActor(id uint64) (IActor, error) {
	if act := d.GetActor(id); act != nil {
		return act, nil
	}

	sh := d.shards[id%uint64(d.attr.GetSize())]
	sh.mutex.Lock()
	defer sh.mutex.RUnlock()

	// 二次获取
	if act, ok := sh.actors[id]; ok {
		return act, nil
	}

	// 初始化
	obj := any(new(T)).(IActor)
	if !obj.Register(obj, d.attr.ToOptions()...) {
		return nil, fmt.Errorf("%s(%d)启动任务队列失败", d.attr.GetName(), id)
	}
	sh.actors[id] = obj
	return obj, nil
}
