package mapstruct

import "sync"

type Group4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[four[T1, T2, T3, T4]][]V

// 大小
func (d Group4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

// 设置
func (d Group4[T1, T2, T3, T4, V]) Put(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	d[key] = append(d[key], value)
}

// 读取
func (d Group4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	value, ok := d[key]
	return value, ok
}

// 删除
func (d Group4[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Group4S[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[four[T1, T2, T3, T4]][]V
}

func NewGroup4S[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any]() *Group4S[T1, T2, T3, T4, V] {
	return &Group4S[T1, T2, T3, T4, V]{data: make(map[four[T1, T2, T3, T4]][]V)}
}

// 大小
func (d *Group4S[T1, T2, T3, T4, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Group4S[T1, T2, T3, T4, V]) Put(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.data[key] = append(d.data[key], value)
}

// 读取
func (d *Group4S[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	values, ok := d.data[key]
	return values, ok
}

// 删除
func (d *Group4S[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) ([]V, bool) {
	key := four[T1, T2, T3, T4]{t1, t2, t3, t4}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	values, ok := d.data[key]
	if ok {
		delete(d.data, key)
	}
	return values, ok
}
