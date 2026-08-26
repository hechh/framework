package actor

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hechh/framework/core/define"
	"github.com/hechh/framework/library/queue"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/library/utils"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Shards struct {
	mutex  sync.RWMutex
	actors map[uint64]IActor
}

func (d *Shards) RLock()   { d.mutex.RLock() }
func (d *Shards) RUnlock() { d.mutex.RUnlock() }
func (d *Shards) Lock()    { d.mutex.Lock() }
func (d *Shards) Unlock()  { d.mutex.Unlock() }

type ActorMgr[T any] struct {
	attr   *queue.Attribute
	shards []*Shards
}

func (d *ActorMgr[T]) GetName() string { return d.attr.GetName() }
func (d *ActorMgr[T]) GetId() uint64   { return d.attr.GetId() }

func (d *ActorMgr[T]) Start() error {
	if d.attr.IsClosed() {
		var list []IActor
		for _, sh := range d.shards {
			sh.mutex.RLock()
			for _, act := range sh.actors {
				if err := act.Start(); err != nil {
					defer sh.mutex.RUnlock()
					for i := len(list) - 1; i >= 0; i-- {
						list[i].Stop()
					}
					return err
				}
				list = append(list, act)
			}
			sh.mutex.RUnlock()
		}
		d.attr.Running()
	}
	return nil
}

func (d *ActorMgr[T]) Stop() {
	if d.attr.IsRunning() {
		for _, sh := range d.shards {
			sh.mutex.Lock()
			for _, act := range sh.actors {
				act.Stop()
			}
			sh.mutex.Unlock()
		}
		d.attr.Waiting()
	}
}

func (d *ActorMgr[T]) Register(ac IActor, c define.ICache, opts ...queue.Option) {
	name := utils.ParseName(reflect.TypeOf(ac))
	opts = append(opts, queue.WithName(name))
	d.attr = new(queue.Attribute)
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
		id := tplutil.Or(head.ActorId > 0, head.ActorId, head.Uid)
		err := d.GetActor(id).SendMsg(head, args...)
		if err != nil {
			mlog.Errorf("ActorId(%d)消息转发失败 error=%v", id, err)
		}
		return err
	case packet.SendType_BROADCAST:
		err := d.BroadcastMsg(head, args...)
		if err != nil {
			mlog.Errorf("%s 广播失败 error=%v", d.GetName(), err)
		}
		return err
	default:
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.attr.GetName(), head.SendType)
	}
}

func (d *ActorMgr[T]) Send(head *packet.Head, body []byte) error {
	switch head.SendType {
	case packet.SendType_POINT:
		id := tplutil.Or(head.ActorId > 0, head.ActorId, head.Uid)
		err := d.GetActor(id).Send(head, body)
		if err != nil {
			mlog.Errorf("ActorId(%d)消息转发失败 error=%v", id, err)
		}
		return err
	case packet.SendType_BROADCAST:
		err := d.Broadcast(head, body)
		if err != nil {
			mlog.Errorf("%s 广播失败 error=%v", d.GetName(), err)
		}
		return err
	default:
		return fmt.Errorf("ActorMgr(%s)不支该发送类型(%d)", d.attr.GetName(), head.SendType)
	}
}

func (d *ActorMgr[T]) BroadcastMsg(head *packet.Head, args ...any) (err error) {
	for _, sh := range d.shards {
		sh.mutex.RLock()
		for _, act := range sh.actors {
			if reterr := act.SendMsg(head, args...); reterr != nil {
				err = reterr
			}
		}
		sh.mutex.RUnlock()
	}
	return
}

func (d *ActorMgr[T]) Broadcast(head *packet.Head, body []byte) (err error) {
	for _, sh := range d.shards {
		sh.mutex.RLock()
		for _, act := range sh.actors {
			if reterr := act.Send(head, body); reterr != nil {
				err = reterr
			}
		}
		sh.mutex.RUnlock()
	}
	return
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

func (d *ActorMgr[T]) AddActor(act IActor) {
	sh := d.shards[act.GetId()%uint64(d.attr.GetSize())]
	sh.mutex.Lock()
	sh.actors[act.GetId()] = act
	sh.mutex.Unlock()
}

func (d *ActorMgr[T]) DelActor(id uint64) {
	sh := d.shards[id%uint64(d.attr.GetSize())]
	sh.mutex.Lock()
	delete(sh.actors, id)
	sh.mutex.Unlock()
}
