package actor

import (
	"framework/core/define"
	"framework/core/global"
	"framework/library/timer"
	"framework/library/uerror"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type ActorMgr struct {
	mutex  sync.RWMutex
	actors map[uint64]define.IActor
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

func (d *ActorMgr) Register(ac define.IActor, counts ...int) {
	d.id = global.GenerateActorId()
	d.name = global.ParseActorName(reflect.TypeOf(ac))
	d.actors = make(map[uint64]define.IActor)
}

func (d *ActorMgr) RegisterTimer(ctx define.IContext, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(ctx); err != nil {
			ctx.Errorf("ActorMgr定时器转发失败:%v", err)
		}
	})
}

func (d *ActorMgr) SendMsg(ctx define.IContext, args ...any) error {
	switch ctx.GetHead().SendType {
	case 0:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			ctx.AddDepth(1)
			return act.SendMsg(ctx, args...)
		}
		return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
	default:
		d.mutex.RLock()
		defer d.mutex.RUnlock()
		for _, act := range d.actors {
			if err := act.SendMsg(ctx, args...); err != nil {
				ctx.Errorf("%s.%s广播失败\t错误:%v", ctx.GetActorName(), ctx.GetFuncName(), err)
			}
		}
		return nil
	}
}

func (d *ActorMgr) Send(ctx define.IContext, buf []byte) error {
	switch ctx.GetHead().SendType {
	case 0:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			ctx.AddDepth(1)
			return act.Send(ctx, buf)
		}
		return uerror.New(-1, "%s.%s未注册", ctx.GetActorName(), ctx.GetFuncName())
	default:
		d.mutex.RLock()
		defer d.mutex.RUnlock()
		for _, act := range d.actors {
			if err := act.SendMsg(ctx, buf); err != nil {
				ctx.Errorf("%s.%s广播失败\t错误:%v", ctx.GetActorName(), ctx.GetFuncName(), err)
			}
		}
		return nil
	}
}

func (d *ActorMgr) Size() int {
	return len(d.actors)
}

func (d *ActorMgr) GetActor(id uint64) define.IActor {
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

func (d *ActorMgr) AddActor(act define.IActor) (ret bool) {
	if ret = atomic.CompareAndSwapUint32(&d.status, 0, 0); ret {
		id := act.GetActorId()
		d.mutex.Lock()
		d.actors[id] = act
		d.mutex.Unlock()
	}
	return
}
