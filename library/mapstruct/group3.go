package mapstruct

import "sync"

type Group3[T1 comparable, T2 comparable, T3 comparable, V any] map[three[T1, T2, T3]][]V

// 大小
func (d Group3[T1, T2, T3, V]) Size() int {
	return len(d)
}

// 设置
func (d Group3[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d[key] = append(d[key], value)
}

// 读取
func (d Group3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Group3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Group3S[T1 comparable, T2 comparable, T3 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[three[T1, T2, T3]][]V
}

func NewGroup3S[T1 comparable, T2 comparable, T3 comparable, V any]() *Group3S[T1, T2, T3, V] {
	return &Group3S[T1, T2, T3, V]{data: make(map[three[T1, T2, T3]][]V)}
}

// 大小
func (d *Group3S[T1, T2, T3, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Group3S[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = append(d.data[key], value)
}

// 读取
func (d *Group3S[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	values, ok := d.data[key]
	return values, ok
}

// 删除
func (d *Group3S[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	values, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	return values, ok
}
