package structure

import "sync"

type Map3s[T1 comparable, T2 comparable, T3 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Triple[T1, T2, T3]]V
}

func NewMap3s[T1 comparable, T2 comparable, T3 comparable, V any]() *Map3s[T1, T2, T3, V] {
	return &Map3s[T1, T2, T3, V]{data: make(map[Triple[T1, T2, T3]]V)}
}

// 大小
func (d *Map3s[T1, T2, T3, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map3s[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	d.mutex.Lock()
	d.data[Triple[T1, T2, T3]{t1, t2, t3}] = value
	d.mutex.Unlock()
}

// 读取
func (d *Map3s[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	d.mutex.RLock()
	value, ok := d.data[Triple[T1, T2, T3]{t1, t2, t3}]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Map3s[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
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

func (d *Map3s[T1, T2, T3, V]) Walk(f func(V) bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, item := range d.data {
		if !f(item) {
			return
		}
	}
}
