package gc

import (
	"sync"

	"github.com/hechh/framework/library/queue"
)

type Gc struct {
	wg       sync.WaitGroup       // 等待所有任务完成
	tasks    *queue.Queue[func()] // 任务队列
	notifyCh chan struct{}        // 通知通道
	exitCh   chan struct{}        // 退出通道
}

func (d *Gc) Init() error {
	d.tasks = queue.NewQueue[func()]()
	d.notifyCh = make(chan struct{}, 1)
	d.exitCh = make(chan struct{})

	d.wg.Add(1)
	go d.run()
	return nil
}

func (d *Gc) Close() {
	close(d.exitCh)
	d.wg.Wait()
}

func (d *Gc) Push(f func()) {
	d.tasks.Push(f, func() {
		select {
		case d.notifyCh <- struct{}{}:
		default:
		}
	})
}

func (d *Gc) run() {
	defer func() {
		for {
			f, ok := d.tasks.Pop()
			if !ok {
				break
			}
			f()
		}
		d.wg.Done()
	}()

	for {
		select {
		case <-d.notifyCh:
			for {
				f, ok := d.tasks.Pop()
				if !ok {
					break
				}
				f()
			}
		case <-d.exitCh:
			return
		}
	}
}
