package gc

import (
	"framework/library/async"
	"sync"
)

var (
	obj = &Collector{
		tasks:  async.NewQueue[func()](),
		notify: make(chan struct{}, 1),
		exit:   make(chan struct{}),
	}
)

type Collector struct {
	sync.WaitGroup
	tasks  *async.Queue[func()]
	notify chan struct{}
	exit   chan struct{}
}

func Init() {
	defer func() {
		for f := obj.tasks.Pop(); f != nil; f = obj.tasks.Pop() {
			async.Recover(f)
		}
		obj.Done()
	}()

	for {
		select {
		case <-obj.notify:
			for f := obj.tasks.Pop(); f != nil; f = obj.tasks.Pop() {
				async.Recover(f)
			}
		case <-obj.exit:
			return
		}
	}
}

func Close() {
	close(obj.exit)
	obj.Wait()
}

func Put(f func()) {
	obj.tasks.Push(f)
	select {
	case obj.notify <- struct{}{}:
	default:
	}
}
