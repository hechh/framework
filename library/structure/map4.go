package structure

import "sync"

type Map4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[Quad[T1, T2, T3, T4]]V

// 大小
func (d Map4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

// 设置
func (d Map4[T1, T2, T3, T4, V]) Set(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d[key] = value
}

// 读取
func (d Map4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Map4[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Map4S[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Quad[T1, T2, T3, T4]]V
}

func NewMap4S[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any]() *Map4S[T1, T2, T3, T4, V] {
	return &Map4S[T1, T2, T3, T4, V]{data: make(map[Quad[T1, T2, T3, T4]]V)}
}

// 大小
func (d *Map4S[T1, T2, T3, T4, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map4S[T1, T2, T3, T4, V]) Set(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = value
}

// 读取
func (d *Map4S[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	value, ok := d.data[key]
	return value, ok
}

// 删除
func (d *Map4S[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	value, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	return value, ok
}
