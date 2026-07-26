package msgqueue

import (
	"sync/atomic"
	"time"

	"github.com/hechh/framework/common/constant"
)

type ITask interface {
	Do()
	GetMask() uint32
}

type IAsync interface {
	GetName() string       // Actor名字
	GetSize() int          // 获取数量
	GetId() uint64         // Actor ID
	SetId(uint64)          // Actor ID
	GetIdPointer() *uint64 // 获取指针
	Start() bool           // 启动actor任务队列协程
	Stop()                 // 关闭actor任务队列协程
	Wait()                 // 等待actor协程退出
	Push(ITask) bool       // 推送任务
}

type Option func(*Attribute)

type AsyncTimer struct {
	name     string
	interval time.Duration
	times    int32
}

type Attribute struct {
	name       string                            // 名字
	id         uint64                            // 唯一id
	status     int32                             // 状态
	size       int                               // 协程池大小
	idleSecond int64                             // 闲置时间（秒）
	timers     []*AsyncTimer                     // 定时器
	deleteFunc func()                            // 删除函数
	lockSecond int64                             // lock有效时长
	lockFunc   func(uint64, time.Duration) error // 全局任务锁
	unlockFunc func(uint64) error                // 全局任务解锁函数
}

func (d *Attribute) GetSize() int {
	return d.size
}

func (d *Attribute) IsRunning() bool {
	return atomic.LoadInt32(&d.status) == constant.RUNNING_STATUS
}

func (d *Attribute) Running() {
	atomic.StoreInt32(&d.status, constant.RUNNING_STATUS)
}

func (d *Attribute) IsWaiting() bool {
	return atomic.LoadInt32(&d.status) == constant.WAITING_STATUS
}

func (d *Attribute) Waiting() {
	atomic.StoreInt32(&d.status, constant.WAITING_STATUS)
}

func (d *Attribute) IsStopped() bool {
	return atomic.LoadInt32(&d.status) == constant.STOPPED_STATUS
}

func (d *Attribute) Stopped() {
	atomic.StoreInt32(&d.status, constant.STOPPED_STATUS)
}

func (d *Attribute) GetName() string {
	return d.name
}

func (d *Attribute) GetIdPointer() *uint64 {
	return &d.id
}

func (d *Attribute) GetId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *Attribute) SetId(val uint64) {
	atomic.StoreUint64(&d.id, val)
}

func (d *Attribute) OnLock() error {
	if d.lockFunc != nil {
		return d.lockFunc(d.GetId(), time.Duration(d.lockSecond)*time.Second)
	}
	return nil
}

func (d *Attribute) OnUnlock() error {
	if d.unlockFunc != nil {
		return d.unlockFunc(d.GetId())
	}
	return nil
}

func (d *Attribute) OnDelete() {
	if d.deleteFunc != nil {
		d.deleteFunc()
	}
}

func WithName(name string) Option {
	return func(opt *Attribute) {
		opt.name = name
	}
}

func WithSize(size int) Option {
	return func(opt *Attribute) {
		opt.size = size
	}
}

func WithIdleTime(idle int64) Option {
	return func(opt *Attribute) {
		opt.idleSecond = idle
	}
}

func WithDeleter(f func()) Option {
	return func(opt *Attribute) {
		opt.deleteFunc = f
	}
}

func WithLocker(expire int64, lock func(uint64, time.Duration) error, unlock func(uint64) error) Option {
	return func(opt *Attribute) {
		opt.lockSecond = expire
		opt.lockFunc = lock
		opt.unlockFunc = unlock
	}
}

func WithTimer(name string, interval time.Duration, times int32) Option {
	return func(opt *Attribute) {
		opt.timers = append(opt.timers, &AsyncTimer{
			name:     name,
			interval: interval,
			times:    times,
		})
	}
}
