package actor

import (
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/uerror"
)

type ActorMgr struct {
	mutex  sync.RWMutex
	actors map[uint64]framework.IActor
	status uint32
	id     uint64
	name   string
}

func (d *ActorMgr) Start() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Start()
	}
}

func (d *ActorMgr) Stop() {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Stop()
	}
}

func (d *ActorMgr) Done() {
	atomic.StoreUint32(&d.status, 1)
}

func (d *ActorMgr) Wait() {}

func (d *ActorMgr) GetActorName() string {
	return d.name
}

func (d *ActorMgr) GetActorId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *ActorMgr) SetActorId(id uint64) {
	atomic.StoreUint64(&d.id, id)
}

func (d *ActorMgr) Register(ac framework.IActor, counts ...int) {
	d.id = framework.GenActorId()
	d.name = framework.ParseActorName(reflect.TypeOf(ac))
	d.actors = make(map[uint64]framework.IActor)
}

func (d *ActorMgr) RegisterTimer(ctx framework.IContext, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(ctx); err != nil {
			ctx.Errorf("ActorMgr定时器转发失败:%v", err)
		}
	})
}

func (d *ActorMgr) SendMsg(ctx framework.IContext, args ...any) error {
	switch ctx.GetHead().SendType {
	case packet.SendType_POINT:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			ctx.AddDepth(1)
			return act.SendMsg(ctx, args...)
		}
		return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
	case packet.SendType_BROADCAST:
		d.mutex.RLock()
		defer d.mutex.RUnlock()
		for _, act := range d.actors {
			if err := act.SendMsg(ctx, args...); err != nil {
				ctx.Errorf("%s广播失败:%v", ctx.GetActorFunc(), err)
			}
		}
	}
	return nil
}

func (d *ActorMgr) Send(ctx framework.IContext, buf []byte) error {
	switch ctx.GetHead().SendType {
	case 0:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			ctx.AddDepth(1)
			return act.Send(ctx, buf)
		}
		return uerror.New(-1, "%s未注册", ctx.GetActorFunc())
	default:
		d.mutex.RLock()
		defer d.mutex.RUnlock()
		for _, act := range d.actors {
			if err := act.SendMsg(ctx, buf); err != nil {
				ctx.Errorf("%s.%s广播失败:%v", ctx.GetActorFunc(), err)
			}
		}
		return nil
	}
}

func (d *ActorMgr) Size() int {
	return len(d.actors)
}

func (d *ActorMgr) GetActor(id uint64) framework.IActor {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if actor, ok := d.actors[id]; ok {
		return actor
	}
	return nil
}

func (d *ActorMgr) DelActor(id uint64) {
	d.mutex.Lock()
	delete(d.actors, id)
	d.mutex.Unlock()
}

func (d *ActorMgr) AddActor(act framework.IActor) (ret bool) {
	if ret = atomic.CompareAndSwapUint32(&d.status, 0, 0); ret {
		id := act.GetActorId()
		d.mutex.Lock()
		d.actors[id] = act
		d.mutex.Unlock()
	}
	return
}
