package mapstruct

import "sync"

type three[T1 comparable, T2 comparable, T3 comparable] struct {
	f1 T1
	f2 T2
	f3 T3
}

type Map3[T1 comparable, T2 comparable, T3 comparable, V any] map[three[T1, T2, T3]]V

// 大小
func (d Map3[T1, T2, T3, V]) Size() int {
	return len(d)
}

// 设置
func (d Map3[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d[key] = value
}

// 读取
func (d Map3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Map3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Map3S[T1 comparable, T2 comparable, T3 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[three[T1, T2, T3]]V
}

func NewMap3S[T1 comparable, T2 comparable, T3 comparable, V any]() *Map3S[T1, T2, T3, V] {
	return &Map3S[T1, T2, T3, V]{data: make(map[three[T1, T2, T3]]V)}
}

// 大小
func (d *Map3S[T1, T2, T3, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map3S[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = value
}

// 读取
func (d *Map3S[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, ok := d.data[key]
	return value, ok
}

// 删除
func (d *Map3S[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := three[T1, T2, T3]{t1, t2, t3}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	return value, ok
}
