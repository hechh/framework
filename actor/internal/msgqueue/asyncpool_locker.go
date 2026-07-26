package msgqueue

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/common/constant"
	"github.com/hechh/framework/core/actor/internal/domain"

	"github.com/hechh/library/base/queue"
	"github.com/hechh/library/base/safe"
	"github.com/hechh/library/mlog"
)

type AsyncPoolLocker struct {
	*domain.Attribute                            // 基础
	tasks             *queue.Queue[domain.ITask] // 任务队列
	notifyCh          chan struct{}              // 通知
	exitCh            chan struct{}              // 退出
	taskCh            chan domain.ITask          // 任务抢占队列
	updateTime        int64                      // 更新时间
	lockTime          int64                      // 更新时间
	w1                sync.WaitGroup             // 等待 run goroutine 退出
	w2                sync.WaitGroup             // 等待 run goroutine 退出
	startWg           sync.WaitGroup             // 启动状态
}

func NewAsyncPoolLocker(base *domain.Attribute) *AsyncPoolLocker {
	return &AsyncPoolLocker{
		Attribute: base,
		tasks:     queue.NewQueue[domain.ITask](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
		taskCh:    make(chan domain.ITask, 5*base.Size),
	}
}

func (d *AsyncPoolLocker) Start() bool {
	if atomic.LoadInt32(&d.Status) != domain.RUNNING_STATUS {
		// 启动任务队列
		d.startWg.Add(1)
		d.w1.Add(1)
		safe.SafeGo(mlog.Fatalf, d.run)
		d.startWg.Wait()

		// 判断是否已经启动
		if atomic.LoadInt32(&d.Status) != domain.WAITING_STATUS {
			return false
		}

		// 启动处理协程
		for range d.Size {
			d.w2.Add(1)
			safe.SafeGo(mlog.Fatalf, func() {
				defer d.w2.Done()
				for task := range d.taskCh {
					task.Do()
				}
			})
		}
		atomic.StoreInt32(&d.Status, domain.RUNNING_STATUS)
	}
	return true
}

func (d *AsyncPoolLocker) Stop() {
	if atomic.LoadInt32(&d.Status) != domain.STOPPED_STATUS {
		close(d.exitCh)
		atomic.StoreInt32(&d.Status, domain.STOPPED_STATUS)
		d.OnDelete()
	}
}

func (d *AsyncPoolLocker) Wait() {
	id := d.GetId()
	d.SetId(0)
	d.w1.Wait()
	close(d.taskCh)
	d.w2.Wait()
	mlog.Infof("%s(%d)关闭成功", d.Name, id)
}

func (d *AsyncPoolLocker) Push(t domain.ITask) (flag bool) {
	flag = atomic.LoadInt32(&d.Status) == domain.RUNNING_STATUS
	if flag {
		d.tasks.Push(t, func() {
			mlog.Tracef("%s任务数量：%d", d.Name, d.tasks.GetCount())
			select {
			case d.notifyCh <- struct{}{}:
			default:
			}
		})
	}
	return flag
}

func (d *AsyncPoolLocker) run() {
	defer d.w1.Done()

	// 先抢占锁
	if err := d.OnLock(); err != nil {
		mlog.Errorf("%s抢占全局锁失败", d.Name)
		return
	}

	// 启动成功
	atomic.StoreInt32(&d.Status, domain.WAITING_STATUS)
	d.startWg.Done()

	// 保活全局锁
	tt := time.NewTicker(time.Second)
	defer func() {
		tt.Stop()
		d.Stop()     // 发送停止消息
		d.handle()   // 处理剩余请求
		d.OnUnlock() // 释放锁
	}()
	d.updateTime = time.Now().Unix()
	d.lockTime = d.updateTime
	expire := d.LockSecond * 2 / 3

	// 循环处理任务
	for {
		select {
		case <-d.notifyCh:
			d.handle()
		case <-d.exitCh:
			return
		case tnow := <-tt.C:
			if d.IdleSecond > 0 && tnow.Unix()-d.updateTime > d.IdleSecond {
				return
			}
			if tnow.Unix()-d.lockTime >= expire {
				if err := d.OnLock(); err != nil {
					return
				}
				d.lockTime = tnow.Unix()
			}
		}
	}
}

func (d *AsyncPoolLocker) handle() {
	for f := d.tasks.Pop(); f != nil; f = d.tasks.Pop() {
		if f.GetMask()&constant.UPDATETIME_MASK != constant.UPDATETIME_MASK {
			d.updateTime = time.Now().Unix()
		}
		d.taskCh <- f
	}
}
