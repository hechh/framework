package queue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/safe"
)

type MsgQueuePool[T ITask] struct {
	*Attribute                // 基础
	tasks      *Queue[T]      // 任务队列
	notifyCh   chan struct{}  // 通知
	exitCh     chan struct{}  // 退出
	taskCh     chan T         // 任务抢占队列
	updateTime int64          // 更新时间
	lockTime   int64          // 更新时间
	w1         sync.WaitGroup // 等待 run goroutine 退出
	w2         sync.WaitGroup // 等待 run goroutine 退出
}

func NewMsgQueuePool[T ITask](opts ...Option) *MsgQueuePool[T] {
	attr := new(Attribute)
	for _, opt := range opts {
		opt(attr)
	}
	if attr.id <= 0 {
		attr.id = GenId()
	}
	return &MsgQueuePool[T]{
		Attribute: attr,
		tasks:     NewQueue[T](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
		taskCh:    make(chan T, 5*attr.size),
	}
}

func (d *MsgQueuePool[T]) Start(except func(string, ...any)) error {
	if d.IsClosed() {
		// 正在启动
		d.Starting()
		// 先抢占锁
		if err := d.OnLock(); err != nil {
			return err
		}
		// 启动协程池
		for range d.size {
			d.w2.Add(1)
			safe.SafeGo(except, func() {
				d.w2.Done()
				for task := range d.taskCh {
					if task.Do() {
						atomic.StoreInt64(&d.updateTime, time.Now().Unix())
					}
				}
			})
		}
		// 启动协程
		d.w1.Add(1)
		safe.SafeGo(except, d.run)
		// 正在运行
		d.Running()
	}
	return nil
}

func (d *MsgQueuePool[T]) Stop() {
	// 原子地将 RUNNING → WAITING，仅切换成功者执行关闭动作。
	// OnDelete 回调可能重入 Stop 或存在并发调用，CAS 保证 exitCh 只 close 一次。
	if atomic.CompareAndSwapInt32(&d.status, RUNNING, WAITING) {
		close(d.exitCh)
		// 删除
		d.OnDelete()
	}
}

func (d *MsgQueuePool[T]) Wait() {
	if d.IsWaiting() {
		d.w1.Wait()
		d.SetId(0)
		close(d.taskCh)
		d.w2.Wait()
		d.Closed()
	}
}

func (d *MsgQueuePool[T]) Push(t T) (flag bool) {
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

func (d *MsgQueuePool[T]) run() {
	tt := time.NewTicker(time.Second)
	defer func() {
		d.w1.Done()
		tt.Stop()    // 关闭定时器
		d.Stop()     // 发送停止消息
		d.handle()   // 处理剩余请求
		d.OnUnlock() // 释放锁
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
			if d.idleTime > 0 && tnow.Unix()-atomic.LoadInt64(&d.updateTime) > d.idleTime {
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

func (d *MsgQueuePool[T]) handle() {
	for range 100 {
		f, ok := d.tasks.Pop()
		if !ok {
			return
		}
		d.taskCh <- f
	}
	if d.tasks.GetCount() > 0 {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	}
}
