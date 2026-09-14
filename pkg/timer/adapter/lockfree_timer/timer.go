package lockfree_timer

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/timer"
)

type Timer struct {
	wheelSize int              // 轮子数量
	wheels    []*Wheel         // 轮子数组
	period    int64            // 最小周期，单位毫秒,必须是2的N次方倍
	offset    int64            // 偏移量
	startTime int64            // 定时器启动时间
	exitCh    chan struct{}    // 退出通道
	taskCh    chan timer.ITask // 任务通道
	wg        sync.WaitGroup   // 等待 handler goroutine 退出
}

func NewTimer() *Timer {
	return new(Timer)
}

// 添加定时任务
func (d *Timer) Register(task timer.ITask) error {
	if task.GetTTL()>>d.offset <= 0 {
		return fmt.Errorf("不能小于最小定时时间限制")
	}
	now := task.GetExpire()
	for _, wheel := range d.wheels {
		bucket := wheel.Get(now)
		if bucket == nil {
			continue
		}
		bucket.Push(task, nil)
		return nil
	}
	return fmt.Errorf("不能超出最大定时时间限制")
}

func (d *Timer) Init(cfg *timer.Config) error {
	d.wheelSize = cfg.Size
	d.wheels = make([]*Wheel, cfg.Size)
	d.period = 1 << cfg.MinPeriodBitNumber
	d.offset = cfg.MinPeriodBitNumber
	d.exitCh = make(chan struct{})
	d.taskCh = make(chan timer.ITask, 100)
	offset := d.offset
	for i := 0; i < cfg.Size; i++ {
		if i == 0 {
			d.wheels[i] = NewWheel(1024, offset)
			offset += 10
		} else {
			d.wheels[i] = NewWheel(512, offset)
			offset += 9
		}
	}

	// 初始化
	nowMs := datetime.NowUnixMilli()
	for _, wheel := range d.wheels {
		wheel.Refresh(nowMs)
	}
	d.startTime = nowMs

	// 启动执行协程
	for range d.wheelSize {
		d.wg.Add(1)
		safe.SafeGo(mlog.Fatalf, d.consume)
	}

	// 启动定时协程
	d.wg.Add(1)
	safe.SafeGo(mlog.Fatalf, d.run)
	return nil
}

func (d *Timer) Close() {
	close(d.exitCh)
	d.wg.Wait()
}

func (d *Timer) run() {
	tt := time.NewTicker(time.Duration(d.period) * time.Millisecond)
	defer func() {
		tt.Stop()
		d.wg.Done()
	}()

	for {
		select {
		case <-d.exitCh:
			return
		case <-tt.C:
			// 使用 tick 计数器避免 bucket 排空耗时导致后续 tick 被跳过
			nowMs := datetime.NowUnixMilli()
			tickCount := (nowMs - d.startTime) / d.period
			for i := int64(1); i <= tickCount; i++ {
				begin := d.startTime + i*d.period
				for _, wheel := range d.wheels {
					bucket := wheel.Get(begin)
					if bucket == nil {
						continue
					}

					// 更新cursor
					wheel.Refresh(begin)

					// 移动任务
					for {
						item, ok := bucket.Pop()
						if !ok {
							break
						}
						if item.GetExpire()-nowMs <= d.period {
							d.taskCh <- item
						} else {
							d.Register(item)
						}
					}
				}
			}
			d.startTime = d.startTime + tickCount*d.period
		}
	}
}

func (d *Timer) consume() {
	defer d.wg.Done()
	for {
		select {
		case <-d.exitCh:
			return
		case item := <-d.taskCh:
			// 任务是否有效
			if !item.IsEnable() {
				continue
			}

			// 指定任务（panic 已内部捕获）
			d.call(item)

			// 判断任务是否
			if item.IsEnable() {
				item.Refresh(datetime.NowUnixMilli())
				d.Register(item)
			}
		}
	}
}

// call 执行任务回调并捕获 panic：回调 panic 若逃逸会终止当前 worker goroutine，
// 每个 panic 永久少一个 worker（size=5 打光后全进程定时器静默停摆），
// 且该任务已从轮中弹出不会重入；捕获后 worker 循环继续。
func (d *Timer) call(item timer.ITask) {
	defer func() {
		if err := recover(); err != nil {
			mlog.Errorf("定时任务执行panic, task:%T, error:%v\n%s", item, err, string(debug.Stack()))
		}
	}()
	item.Call()
}
