package structure

import "sync"

type Group3s[T1 comparable, T2 comparable, T3 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Triple[T1, T2, T3]][]V
}

func NewGroup3s[T1 comparable, T2 comparable, T3 comparable, V any]() *Group3s[T1, T2, T3, V] {
	return &Group3s[T1, T2, T3, V]{data: make(map[Triple[T1, T2, T3]][]V)}
}

// 大小
func (d *Group3s[T1, T2, T3, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Group3s[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	d.mutex.Lock()
	d.data[key] = append(d.data[key], value)
	d.mutex.Unlock()
}

// 读取
func (d *Group3s[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	d.mutex.RLock()
	values, ok := d.data[key]
	d.mutex.RUnlock()
	return values, ok
}

// 删除
func (d *Group3s[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
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
