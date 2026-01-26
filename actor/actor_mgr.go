package actor

import (
	"myplay/common/pb"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/context"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
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
	atomic.StoreUint32(&d.status, 1)
}

func (d *ActorMgr) Stop() {
	atomic.StoreUint32(&d.status, 0)
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Stop()
	}
}

func (d *ActorMgr) Done() {
	atomic.StoreUint32(&d.status, 0)
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		act.Done()
	}
}

func (d *ActorMgr) Wait() {
	atomic.StoreUint64(&d.id, 0)
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, act := range d.actors {
		id := act.GetActorId()
		act.Wait()
		mlog.Infof("%s(%d)关闭成功", d.name, id)
	}
}

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

func (d *ActorMgr) RegisterTimer(name string, ms time.Duration, times int32) error {
	return timer.Register(&d.id, ms, times, func() {
		if err := d.SendMsg(context.NewSimpleContext(d.GetActorId(), name)); err != nil {
			mlog.Errorf("ActorMgr定时器转发失败:%v", err)
		}
	})
}

func (d *ActorMgr) SendMsg(ctx framework.IContext, args ...any) error {
	var err error
	head := ctx.GetHead()
	switch head.SendType {
	case packet.SendType_POINT:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			err = act.SendMsg(ctx, args...)
		} else {
			err = uerror.Err(pb.ErrorCode_ActorIdNotExist, "ActorId(%d)不存在", ctx.GetActorId())
		}
	case packet.SendType_BROADCAST:
		if head.ActorId > 0 {
			if act := d.GetActor(head.ActorId); act != nil {
				err = act.SendMsg(ctx, args...)
			} else {
				err = uerror.Err(pb.ErrorCode_ActorIdNotExist, "ActorId(%d)不存在", head.ActorId)
			}
		} else {
			d.mutex.RLock()
			defer d.mutex.RUnlock()
			for _, act := range d.actors {
				if err = act.SendMsg(ctx, args...); err != nil {
					break
				}
			}
		}
	}
	mlog.Trace(-1, "[actor] ActorMgr(%s)本地调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, args)
	return err
}

func (d *ActorMgr) Send(ctx framework.IContext, buf []byte) error {
	var err error
	head := ctx.GetHead()
	switch head.SendType {
	case packet.SendType_POINT:
		if act := d.GetActor(ctx.GetActorId()); act != nil {
			err = act.Send(ctx, buf)
		} else {
			err = uerror.Err(pb.ErrorCode_ActorIdNotExist, "ActorId(%d)不存在", ctx.GetActorId())
		}
	case packet.SendType_BROADCAST:
		if head.ActorId > 0 {
			if act := d.GetActor(head.ActorId); act != nil {
				err = act.Send(ctx, buf)
			} else {
				err = uerror.Err(pb.ErrorCode_ActorIdNotExist, "ActorId(%d)不存在", head.ActorId)
			}
		} else {
			d.mutex.RLock()
			defer d.mutex.RUnlock()
			for _, act := range d.actors {
				if err = act.Send(ctx, buf); err != nil {
					break
				}
			}
		}
	}
	mlog.Trace(-1, "[actor] ActorMgr(%s)远程调用 head:%v, error:%v, args:%v", ctx.GetActorFunc(), ctx.GetHead(), err, buf)
	return err
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
	if ret = atomic.CompareAndSwapUint32(&d.status, 1, 1); ret {
		id := act.GetActorId()
		d.mutex.Lock()
		d.actors[id] = act
		d.mutex.Unlock()
	}
	return
}
