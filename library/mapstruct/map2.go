package mapstruct

import "sync"

type two[T1 comparable, T2 comparable] struct {
	f1 T1
	f2 T2
}

type Map2[T1 comparable, T2 comparable, V any] map[two[T1, T2]]V

// 大小
func (d Map2[T1, T2, V]) Size() int {
	return len(d)
}

// 设置
func (d Map2[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	key := two[T1, T2]{t1, t2}
	d[key] = value
}

// 读取
func (d Map2[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	key := two[T1, T2]{t1, t2}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Map2[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	key := two[T1, T2]{t1, t2}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Map2S[T1 comparable, T2 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[two[T1, T2]]V
}

func NewMap2S[T1 comparable, T2 comparable, V any]() *Map2S[T1, T2, V] {
	return &Map2S[T1, T2, V]{data: make(map[two[T1, T2]]V)}
}

// 大小
func (d *Map2S[T1, T2, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map2S[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	key := two[T1, T2]{t1, t2}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = value
}

// 读取
func (d *Map2S[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	key := two[T1, T2]{t1, t2}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, ok := d.data[key]
	return value, ok
}

// 删除
func (d *Map2S[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	key := two[T1, T2]{t1, t2}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	return value, ok
}
