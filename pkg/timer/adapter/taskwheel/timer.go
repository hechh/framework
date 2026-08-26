package taskwheel

import (
	"sync/atomic"
	"time"

	"github.com/ankur-anand/taskwheel"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/pkg/timer"
)

type taskItem struct {
	task timer.ITask
	id   uint64
}

var globalSeq uint64

type Timer struct {
	wheel  *taskwheel.HierarchicalTimingWheel[*taskItem]
	stopFn func()
	closed int32
}

func NewTimer() *Timer {
	return &Timer{}
}

func (d *Timer) Init(cfg *timer.Config) error {
	// 配置分层时间轮，与 lockfree_timer 的能力范围相近
	// Level 0: 1ms 间隔 × 100 槽 = 100ms 跨度
	// Level 1: 10ms 间隔 × 100 槽 = 10s 跨度
	// Level 2: 100ms 间隔 × 100 槽 = 1000s 跨度
	// Level 3: 1s 间隔 × 60 槽 = 60000s 跨度
	intervals := []time.Duration{1 * time.Millisecond, 10 * time.Millisecond, 100 * time.Millisecond, 1 * time.Second}
	slots := []int{100, 100, 100, 60}
	d.wheel = taskwheel.NewHierarchicalTimingWheel[*taskItem](intervals, slots)

	d.stopFn = d.wheel.StartBatch(1*time.Millisecond, func(ts []*taskwheel.Timer[*taskItem]) {
		for _, t := range ts {
			if t == nil || t.Value == nil {
				continue
			}
			item := t.Value
			task := item.task

			// 执行任务回调
			task.Call()

			// 如果任务仍有效，重新注册
			if task.IsEnable() {
				now := datetime.NowUnixMilli()
				task.Refresh(now)
				newID := atomic.AddUint64(&item.id, 1)
				_, _ = d.wheel.AfterTimeout(
					taskwheel.TimerID(newID),
					item,
					time.Duration(task.GetTTL())*time.Millisecond,
				)
			}
		}
	})
	return nil
}

func (d *Timer) Close() {
	if atomic.CompareAndSwapInt32(&d.closed, 0, 1) {
		if d.stopFn != nil {
			d.stopFn()
		}
	}
}

func (d *Timer) Register(task timer.ITask) error {
	id := atomic.AddUint64(&globalSeq, 1)
	item := &taskItem{
		task: task,
		id:   id,
	}
	_, err := d.wheel.AfterTimeout(
		taskwheel.TimerID(id),
		item,
		time.Duration(task.GetTTL())*time.Millisecond,
	)
	return err
}
