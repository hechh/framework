package domain

import (
	"sync/atomic"
	"time"
)

type Option func(*Attribute)

type AsyncTimer struct {
	Name     string
	Interval time.Duration
	Times    int32
}

type Attribute struct {
	Name       string                            // 名字
	Id         uint64                            // 唯一id
	Status     int32                             // 状态
	Size       int                               // 协程池大小
	IdleSecond int64                             // 闲置时间（秒）
	Timers     []*AsyncTimer                     // 定时器
	DeleteFunc func()                            // 删除函数
	LockSecond int64                             // lock有效时长
	LockFunc   func(uint64, time.Duration) error // 全局任务锁
	UnlockFunc func(uint64) error                // 全局任务解锁函数
}

func (d *Attribute) GetSize() int {
	return d.Size
}

func (d *Attribute) GetName() string {
	return d.Name
}

func (d *Attribute) GetIdPointer() *uint64 {
	return &d.Id
}

func (d *Attribute) GetId() uint64 {
	return atomic.LoadUint64(&d.Id)
}

func (d *Attribute) SetId(val uint64) {
	atomic.StoreUint64(&d.Id, val)
}

func (d *Attribute) OnLock() error {
	if d.LockFunc != nil {
		return d.LockFunc(d.GetId(), time.Duration(d.LockSecond)*time.Second)
	}
	return nil
}

func (d *Attribute) OnUnlock() error {
	if d.UnlockFunc != nil {
		return d.UnlockFunc(d.GetId())
	}
	return nil
}

func (d *Attribute) OnDelete() {
	if d.DeleteFunc != nil {
		d.DeleteFunc()
	}
}
