package structure

import "sync"

type Group2[T1 comparable, T2 comparable, V any] map[Pair[T1, T2]][]V

func NewGroup2[T1 comparable, T2 comparable, V any]() Group2[T1, T2, V] {
	return make(map[Pair[T1, T2]][]V)
}

// 大小
func (d Group2[T1, T2, V]) Size() int {
	return len(d)
}

// 设置
func (d Group2[T1, T2, V]) Put(t1 T1, t2 T2, value V) {
	key := Pair[T1, T2]{t1, t2}
	d[key] = append(d[key], value)
}

// 读取
func (d Group2[T1, T2, V]) Get(t1 T1, t2 T2) ([]V, bool) {
	key := Pair[T1, T2]{t1, t2}
	value, ok := d[key]
	return value, ok
}

func (d Group2[T1, T2, V]) Copy(t1 T1, t2 T2, f func(V) V) (rets []V, ok bool) {
	if values, ok := d[Pair[T1, T2]{t1, t2}]; ok {
		rets = make([]V, len(values))
		for i, item := range values {
			rets[i] = f(item)
		}
	}
	return
}

// 删除
func (d Group2[T1, T2, V]) Del(t1 T1, t2 T2) ([]V, bool) {
	key := Pair[T1, T2]{t1, t2}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Group2s[T1 comparable, T2 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Pair[T1, T2]][]V
}

func NewGroup2s[T1 comparable, T2 comparable, V any]() *Group2s[T1, T2, V] {
	return &Group2s[T1, T2, V]{data: make(map[Pair[T1, T2]][]V)}
}

// 大小
func (d *Group2s[T1, T2, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Group2s[T1, T2, V]) Put(t1 T1, t2 T2, value V) {
	key := Pair[T1, T2]{t1, t2}
	d.mutex.Lock()
	d.data[key] = append(d.data[key], value)
	d.mutex.Unlock()
}

// 读取
func (d *Group2s[T1, T2, V]) Get(t1 T1, t2 T2) ([]V, bool) {
	key := Pair[T1, T2]{t1, t2}
	d.mutex.RLock()
	value, ok := d.data[key]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Group2s[T1, T2, V]) Del(t1 T1, t2 T2) ([]V, bool) {
	key := Pair[T1, T2]{t1, t2}
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
