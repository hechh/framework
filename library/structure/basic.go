package structure

import "sync"

type Pair[T any, U any] struct {
	First  T
	Second U
}

type Triple[A any, B any, C any] struct {
	First  A
	Second B
	Three  C
}

type Quad[A any, B any, C any, D any] struct {
	First  A
	Second B
	Three  C
	Four   D
}

type Locker[T any] struct {
	mutex sync.RWMutex
	t     T
}

func NewLocker[T any](t T) *Locker[T] {
	return &Locker[T]{t: t}
}

func (d *Locker[T]) Lock() T {
	d.mutex.Lock()
	return d.t
}

func (d *Locker[T]) Unlock() {
	d.mutex.Unlock()
}

func (d *Locker[T]) RLock() T {
	d.mutex.RLock()
	return d.t
}

func (d *Locker[T]) RUnlock() {
	d.mutex.RUnlock()
}
