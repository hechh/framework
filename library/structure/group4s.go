package structure

import "sync"

type Group4s[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Quad[T1, T2, T3, T4]][]V
}

func NewGroup4s[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any]() *Group4s[T1, T2, T3, T4, V] {
	return &Group4s[T1, T2, T3, T4, V]{data: make(map[Quad[T1, T2, T3, T4]][]V)}
}

// 大小
func (d *Group4s[T1, T2, T3, T4, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Group4s[T1, T2, T3, T4, V]) Put(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.Lock()
	d.data[key] = append(d.data[key], value)
	d.mutex.Unlock()
}

// 读取
func (d *Group4s[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.RLock()
	values, ok := d.data[key]
	d.mutex.RUnlock()
	return values, ok
}

// 删除
func (d *Group4s[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.RLock()
	values, ok := d.data[key]
	d.mutex.RUnlock()
	if ok {
		d.mutex.Lock()
		delete(d.data, key)
		d.mutex.Unlock()
	}
	return values, ok
}
