package queue

import (
	"sync"
	"sync/atomic"
)

type node[T any] struct {
	next  atomic.Pointer[node[T]]
	value T
}

type Queue[T any] struct {
	head  atomic.Pointer[node[T]] // 头指针
	tail  atomic.Pointer[node[T]] // 尾指针
	pool  sync.Pool               // *node[T] 对象池
	count atomic.Int32            // 数量
}

func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{}
	q.pool.New = func() any { return new(node[T]) }
	// 哨兵节点也从池中分配
	nn := q.pool.Get().(*node[T])
	nn.next.Store(nil)
	q.head.Store(nn)
	q.tail.Store(nn)
	return q
}

func (d *Queue[T]) GetCount() int32 {
	return d.count.Load()
}

func (d *Queue[T]) Push(val T, cb func()) int32 {
	addNode := d.pool.Get().(*node[T])
	addNode.value = val
	addNode.next.Store(nil)
	prevNode := d.tail.Swap(addNode)
	prevNode.next.Store(addNode)
	// 通知
	count := d.count.Add(1)
	if cb != nil {
		cb()
	}
	return count
}

// Pop 尝试从队列弹出元素，如果队列为空或 CAS 失败则返回零值和 false
func (d *Queue[T]) Pop() (ret T, ok bool) {
	head := d.head.Load()
	next := head.next.Load()
	if next == nil {
		return
	}
	if !d.head.CompareAndSwap(head, next) {
		return
	}
	d.count.Add(-1)
	ret = next.value
	d.pool.Put(head)
	return ret, true
}
