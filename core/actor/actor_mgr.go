package actor

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hechh/framework/define"
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

func (d *ActorMgr[T]) GetName() string { return d.attr.GetName() }
func (d *ActorMgr[T]) GetId() uint64   { return d.attr.GetId() }

func (d *ActorMgr[T]) Start() bool {
	if d.attr.IsStopped() {
		var list []IActor
		for _, sh := range d.shards {
			sh.mutex.RLock()
			for _, act := range sh.actors {
				if act.Start() {
					list = append(list, act)
				} else {
					for i := len(list) - 1; i >= 0; i-- {
						list[i].Stop()
					}
					return false
				}
			}
			sh.mutex.RUnlock()
		}
		d.attr.Running()
	}
	return d.attr.IsRunning()
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

func (d *ActorMgr[T]) Register(ac IActor, c define.ICache, opts ...msgqueue.Option) {
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
		err := d.GetActor(id).SendMsg(head, args...)
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
		err := d.GetActor(id).Send(head, body)
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
