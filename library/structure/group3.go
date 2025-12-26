package structure

import "sync"

type Group3[T1 comparable, T2 comparable, T3 comparable, V any] map[Triple[T1, T2, T3]][]V

func NewGroup3[T1 comparable, T2 comparable, T3 comparable, V any]() Group3[T1, T2, T3, V] {
	return make(map[Triple[T1, T2, T3]][]V)
}

// 大小
func (d Group3[T1, T2, T3, V]) Size() int {
	return len(d)
}

// 设置
func (d Group3[T1, T2, T3, V]) Put(t1 T1, t2 T2, t3 T3, value V) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	d[key] = append(d[key], value)
}

// 读取
func (d Group3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	return value, ok
}

func (d Group3[T1, T2, T3, V]) Copy(t1 T1, t2 T2, t3 T3, f func(V) V) (rets []V, ok bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	if values, ok := d[key]; ok {
		rets = make([]V, len(values))
		for i, item := range values {
			rets[i] = f(item)
		}
	}
	return
}

// 删除
func (d Group3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) ([]V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

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
