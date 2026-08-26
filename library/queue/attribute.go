package queue

import (
	"sync/atomic"
	"time"
)

const (
	STOPPED  = 0 // 已停止
	STARTING = 1 // 启动中
	RUNNING  = 2 // 运行中
	WAITING  = 3 // 等待关闭中
)

var (
	msgqueueId uint64
)

type Attribute struct {
	name       string                            // 名字
	id         uint64                            // 唯一id
	status     int32                             // 状态
	size       int                               // 分片或协程池大小
	idleTime   int64                             // 闲置时间（秒）
	expireTime int64                             // lock有效时长
	locker     func(uint64, time.Duration) error // 全局任务锁
	unlocker   func(uint64) error                // 全局任务解锁函数
	deleter    func(uint64)                      // 删除函数
	//	callback   func(int32, error)                // 回调函数
}

type Option func(*Attribute)

func GenId() uint64 {
	return atomic.AddUint64(&msgqueueId, 1)
}

func WithName(name string) Option {
	return func(opt *Attribute) {
		opt.name = name
	}
}

// WithId 设置唯一id（不修改 name）
func WithId(id uint64) Option {
	return func(opt *Attribute) {
		opt.id = id
	}
}

// WithSize 设置分片大小
func WithSize(size int) Option {
	return func(opt *Attribute) {
		opt.size = size
	}
}

func WithIdleTime(idle int64) Option {
	return func(opt *Attribute) {
		opt.idleTime = idle
	}
}

func WithLocker(expire int64, lock func(uint64, time.Duration) error, unlock func(uint64) error) Option {
	return func(opt *Attribute) {
		opt.expireTime = expire
		opt.locker = lock
		opt.unlocker = unlock
	}
}

func WithDeleter(f func(uint64)) Option {
	return func(opt *Attribute) {
		opt.deleter = f
	}
}

/*
func WithCallback(f func(int32, error)) Option {
	return func(opt *Attribute) {
		opt.callback = f
	}
}
*/

func (d *Attribute) GetSize() int {
	return d.size
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

func (d *Attribute) GetStatus() int32 {
	return atomic.LoadInt32(&d.status)
}

func (d *Attribute) SetStatus(val int32) {
	atomic.StoreInt32(&d.status, val)
}

func (d *Attribute) IsStarting() bool {
	return atomic.LoadInt32(&d.status) == STARTING
}
func (d *Attribute) IsRunning() bool {
	return atomic.LoadInt32(&d.status) == RUNNING
}
func (d *Attribute) IsWaiting() bool {
	return atomic.LoadInt32(&d.status) == WAITING
}
func (d *Attribute) IsClosed() bool {
	return atomic.LoadInt32(&d.status) == STOPPED
}

func (d *Attribute) Starting() {
	atomic.StoreInt32(&d.status, STARTING)
}
func (d *Attribute) Running() {
	atomic.StoreInt32(&d.status, RUNNING)
}
func (d *Attribute) Waiting() {
	atomic.StoreInt32(&d.status, WAITING)
}
func (d *Attribute) Closed() {
	atomic.StoreInt32(&d.status, STOPPED)
}

func (d *Attribute) OnLock() error {
	if d.locker != nil {
		return d.locker(d.GetId(), time.Duration(d.expireTime)*time.Second)
	}
	return nil
}

func (d *Attribute) OnUnlock() error {
	if d.unlocker != nil {
		return d.unlocker(d.GetId())
	}
	return nil
}

func (d *Attribute) OnDelete() {
	if d.deleter != nil {
		d.deleter(d.GetId())
	}
}
