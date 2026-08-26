package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
)

type ITask interface {
	Do() bool
}

type MsgQueue[T ITask] struct {
	*Attribute                // 基础
	tasks      *Queue[T]      // 任务队列
	notifyCh   chan struct{}  // 通知
	exitCh     chan struct{}  // 退出
	updateTime int64          // 更新时间
	lockTime   int64          // 全局锁保活
	wg         sync.WaitGroup // 等待任务完成
}

func NewMsgQueue[T ITask](opts ...Option) *MsgQueue[T] {
	attr := new(Attribute)
	for _, opt := range opts {
		opt(attr)
	}
	if attr.id <= 0 {
		attr.id = GenId()
	}
	return &MsgQueue[T]{
		Attribute: attr,
		tasks:     NewQueue[T](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
	}
}

func (d *MsgQueue[T]) Start(except func(string, ...any)) error {
	if d.IsClosed() {
		// 正在启动
		d.Starting()
		// 先抢占锁
		if err := d.OnLock(); err != nil {
			return err
		}
		// 启动协程
		d.wg.Add(1)
		safe.SafeGo(except, d.run)
		// 正在运行
		d.Running()
	}
	return nil
}

func (d *MsgQueue[T]) Stop() {
	// 原子地将 RUNNING → WAITING，仅切换成功者执行关闭动作。
	// OnDelete 回调（如 PlayerMgr.remove → Actor.Stop）可能在 Waiting 之前
	// 重入 Stop，或存在并发调用；CAS 保证 exitCh 只被 close 一次，
	// 避免 "close of closed channel" panic。
	if atomic.CompareAndSwapInt32(&d.status, RUNNING, WAITING) {
		close(d.exitCh)
		// 删除
		d.OnDelete()
	}
}

func (d *MsgQueue[T]) Wait() {
	if d.IsWaiting() {
		// 等待结束
		d.wg.Wait()
		// 关闭定时器
		d.SetId(0)
		// 修改状态
		d.Closed()
	}
}

func (d *MsgQueue[T]) Push(t T) (flag bool) {
	if flag = d.IsRunning(); flag {
		d.tasks.Push(t, func() {
			select {
			case d.notifyCh <- struct{}{}:
			default:
			}
		})
	}
	return flag
}

func (d *MsgQueue[T]) run() {
	tt := time.NewTicker(time.Second)
	defer func() {
		d.wg.Done()
		tt.Stop()
		d.Stop()
		d.handle()
		d.OnUnlock()
	}()

	d.updateTime = time.Now().Unix()
	d.lockTime = d.updateTime
	lockExpire := d.expireTime * 2 / 3

	// 循环处理任务
	for {
		select {
		case <-d.notifyCh:
			d.handle()
		case <-d.exitCh:
			return
		case tnow := <-tt.C:
			if d.idleTime > 0 && tnow.Unix()-d.updateTime > d.idleTime {
				return
			}
			if d.expireTime > 0 && tnow.Unix()-d.lockTime >= lockExpire {
				if err := d.OnLock(); err != nil {
					return
				}
				d.lockTime = tnow.Unix()
			}
		}
	}
}

func (d *MsgQueue[T]) handle() {
	for range 100 {
		f, ok := d.tasks.Pop()
		if !ok {
			return
		}
		if f.Do() {
			d.updateTime = time.Now().Unix()
		}
	}
	if d.tasks.GetCount() > 0 {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	}
}
