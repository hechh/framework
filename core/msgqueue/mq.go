package msgqueue

import (
	"sync"
	"time"

	"github.com/hechh/framework/common/constant"
	"github.com/hechh/library/base/queue"
	"github.com/hechh/library/base/safe"
	"github.com/hechh/library/mlog"
)

type Async struct {
	*Attribute                     // 基础
	tasks      *queue.Queue[ITask] // 任务队列
	notifyCh   chan struct{}       // 通知
	exitCh     chan struct{}       // 退出
	updateTime int64               // 更新时间
	wg         sync.WaitGroup      // 等待任务完成
	startWg    sync.WaitGroup      // 启动状态
}

func NewAsync(base *Attribute) *Async {
	return &Async{
		Attribute: base,
		tasks:     queue.NewQueue[ITask](),
		notifyCh:  make(chan struct{}, 1),
		exitCh:    make(chan struct{}),
	}
}

func (d *Async) Start() bool {
	if !d.IsRunning() {
		// 启动任务队列
		d.startWg.Add(1)
		d.wg.Add(1)
		safe.SafeGo(mlog.Fatalf, d.run)
		d.startWg.Wait()
		// 判断是否已经启动
		if !d.IsWaiting() {
			return false
		}
		// 启动
		d.Running()
	}
	return true
}

func (d *Async) Stop() {
	if !d.IsStopped() {
		close(d.exitCh)
		d.Stopped()
		d.OnDelete()
	}
}

func (d *Async) Wait() {
	id := d.GetId()
	d.SetId(0)
	d.wg.Wait()
	mlog.Infof("%s(%d)关闭成功", d.name, id)
}

func (d *Async) Push(t ITask) (flag bool) {
	flag = d.IsRunning()
	if flag {
		d.tasks.Push(t, func() {
			mlog.Tracef("%s任务数量：%d", d.name, d.tasks.GetCount())
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
	d.Waiting()
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
			if d.idleSecond > 0 && tnow.Unix()-d.updateTime > d.idleSecond {
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
