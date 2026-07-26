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

type Async struct {
	*domain.Attribute                            // 基础
	tasks             *queue.Queue[domain.ITask] // 任务队列
	notifyCh          chan struct{}              // 通知
	exitCh            chan struct{}              // 退出
	updateTime        int64                      // 更新时间
	wg                sync.WaitGroup             // 等待任务完成
	startWg           sync.WaitGroup             // 启动状态
}

func NewAsync(base *domain.Attribute) *Async {
	return &Async{
		Attribute: base,
		tasks:     queue.NewQueue[domain.ITask](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
	}
}

func (d *Async) Start() bool {
	if atomic.LoadInt32(&d.Status) != domain.RUNNING_STATUS {
		// 启动任务队列
		d.startWg.Add(1)
		d.wg.Add(1)
		safe.SafeGo(mlog.Fatalf, d.run)
		d.startWg.Wait()
		// 判断是否已经启动
		if atomic.LoadInt32(&d.Status) != domain.WAITING_STATUS {
			return false
		}
		// 启动
		atomic.StoreInt32(&d.Status, domain.RUNNING_STATUS)
	}
	return true
}

func (d *Async) Stop() {
	if atomic.LoadInt32(&d.Status) != domain.STOPPED_STATUS {
		close(d.exitCh)
		atomic.StoreInt32(&d.Status, domain.STOPPED_STATUS)
		d.OnDelete()
	}
}

func (d *Async) Wait() {
	id := d.GetId()
	d.SetId(0)
	d.wg.Wait()
	mlog.Infof("%s(%d)关闭成功", d.Name, id)
}

func (d *Async) Push(t domain.ITask) (flag bool) {
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

func (d *Async) run() {
	defer d.wg.Done()

	// 启动成功
	atomic.StoreInt32(&d.Status, domain.WAITING_STATUS)
	d.startWg.Done()

	// 保活全局锁
	tt := time.NewTicker(time.Second)
	defer func() {
		tt.Stop()
		d.Stop()   // 发送停止消息
		d.handle() // 处理剩余请求
	}()
	d.updateTime = time.Now().Unix()

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
		}
	}
}

func (d *Async) handle() {
	for f := d.tasks.Pop(); f != nil; f = d.tasks.Pop() {
		if f.GetMask()&constant.UPDATETIME_MASK != constant.UPDATETIME_MASK {
			d.updateTime = time.Now().Unix()
		}
		f.Do()
	}
}
