package async

import (
	"sync/atomic"
	"unsafe"
)

type node[T any] struct {
	next  *node[T]
	value T
}

type Queue[T any] struct {
	head  *node[T]
	tail  *node[T]
	count int32
}

func NewQueue[T any]() *Queue[T] {
	nn := new(node[T])
	return &Queue[T]{head: nn, tail: nn}
}

func (d *Queue[T]) GetCount() int32 {
	return atomic.LoadInt32(&d.count)
}

func (d *Queue[T]) Push(val T) {
	addNode := new(node[T])
	addNode.value = val
	prevNode := (*node[T])(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&d.tail)), unsafe.Pointer(addNode)))
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&prevNode.next)), unsafe.Pointer(addNode))
	atomic.AddInt32(&d.count, 1)
}

func (d *Queue[T]) Pop() (ret T) {
	if node := (*node[T])(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&d.head.next)))); node != nil {
		atomic.AddInt32(&d.count, -1)
		ret = node.value
		d.head.next = nil
		d.head = node
	}
	return
}
