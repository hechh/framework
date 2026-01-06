package async

import (
	"sync"
	"sync/atomic"
)

type Async struct {
	sync.WaitGroup
	queue  *Queue[func()] // 任务队列
	notify chan struct{}  // 通知
	exit   chan struct{}  // 退出
	id     uint64         // 唯一id
	status int32          // 状态
}

func NewAsync() *Async {
	return &Async{
		queue:  NewQueue[func()](),
		notify: make(chan struct{}, 1),
		exit:   make(chan struct{}),
	}
}

func (d *Async) GetIdPointer() *uint64 {
	return &d.id
}

func (d *Async) GetId() uint64 {
	return atomic.LoadUint64(&d.id)
}

func (d *Async) SetId(id uint64) {
	atomic.StoreUint64(&d.id, id)
}

func (d *Async) Start() {
	if atomic.CompareAndSwapInt32(&d.status, 0, 1) {
		d.Add(1)
		go d.run()
	}
}

func (d *Async) Stop() {
	if atomic.CompareAndSwapInt32(&d.status, 1, 0) {
		close(d.exit)
		d.Wait()
		atomic.StoreUint64(&d.id, 0)
	}
}

func (d *Async) Push(f func()) {
	if atomic.CompareAndSwapInt32(&d.status, 1, 1) {
		d.queue.Push(f)
		select {
		case d.notify <- struct{}{}:
		default:
		}
	}
}

func (d *Async) run() {
	defer func() {
		for f := d.queue.Pop(); f != nil; f = d.queue.Pop() {
			Recover(f)
		}
		d.Done()
	}()
	for {
		select {
		case <-d.notify:
			for f := d.queue.Pop(); f != nil; f = d.queue.Pop() {
				Recover(f)
			}
		case <-d.exit:
			return
		}
	}
}
