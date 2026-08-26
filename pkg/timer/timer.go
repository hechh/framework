package timer

import (
	"fmt"
)

type Config struct {
	Size               int   `yaml:"size"`                  // 时间轮数量
	MinPeriodBitNumber int64 `yaml:"min_period_bit_number"` // 定时器最小周期
}

type ITask interface {
	IsEnable() bool   // 是否可用
	GetTTL() int64    // 获取有效时长
	GetExpire() int64 // 任务触发时间点
	Refresh(int64)    // 刷新任务触发时间点
	Call()            // 执行任务
}

type ITimer interface {
	Init(*Config) error
	Close()
	Register(ITask) error
}

type Timer struct {
	inner  ITimer
	newFun func() ITimer
}

func NewTimer[T ITimer](f func() T) *Timer {
	return &Timer{
		newFun: func() ITimer { return f() },
	}
}

func (d *Timer) Register(task ITask) error {
	if d.inner == nil {
		return fmt.Errorf("定时器未初始化")
	}
	return d.inner.Register(task)
}

func (d *Timer) Init(cfg *Config) error {
	obj := d.newFun()
	if err := obj.Init(cfg); err != nil {
		return err
	}
	d.inner = obj
	return nil
}

func (d *Timer) Close() {
	if d.inner != nil {
		d.inner.Close()
	}
}
