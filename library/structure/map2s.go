package structure

import "sync"

type Map2s[T1 comparable, T2 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Pair[T1, T2]]V
}

func NewMap2s[T1 comparable, T2 comparable, V any]() *Map2s[T1, T2, V] {
	return &Map2s[T1, T2, V]{data: make(map[Pair[T1, T2]]V)}
}

// 大小
func (d *Map2s[T1, T2, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map2s[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	d.mutex.Lock()
	d.data[Pair[T1, T2]{t1, t2}] = value
	d.mutex.Unlock()
}

// 读取
func (d *Map2s[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	d.mutex.RLock()
	value, ok := d.data[Pair[T1, T2]{t1, t2}]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Map2s[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	key := Pair[T1, T2]{t1, t2}
	d.mutex.RLock()
	value, ok := d.data[key]
	d.mutex.RUnlock()
	if ok {
		d.mutex.Lock()
		delete(d.data, key)
		d.mutex.Unlock()
	}
	return value, ok
}

func (d *Map2s[T1, T2, V]) Walk(f func(V) bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, item := range d.data {
		if !f(item) {
			return
		}
	}
}
