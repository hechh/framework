package structure

import "sync"

type Map2S[T1 comparable, T2 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Pair[T1, T2]]V
}

func NewMap2S[T1 comparable, T2 comparable, V any]() *Map2S[T1, T2, V] {
	return &Map2S[T1, T2, V]{data: make(map[Pair[T1, T2]]V)}
}

// 大小
func (d *Map2S[T1, T2, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map2S[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	d.mutex.Lock()
	d.data[Pair[T1, T2]{t1, t2}] = value
	d.mutex.Unlock()
}

// 读取
func (d *Map2S[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	d.mutex.RLock()
	value, ok := d.data[Pair[T1, T2]{t1, t2}]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Map2S[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
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

func (d *Map2S[T1, T2, V]) Walk(f func(V) bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, item := range d.data {
		if !f(item) {
			return
		}
	}
}

type Map3S[T1 comparable, T2 comparable, T3 comparable, V any] struct {
	mutex sync.RWMutex
	data  map[Triple[T1, T2, T3]]V
}

func NewMap3S[T1 comparable, T2 comparable, T3 comparable, V any]() *Map3S[T1, T2, T3, V] {
	return &Map3S[T1, T2, T3, V]{data: make(map[Triple[T1, T2, T3]]V)}
}

// 大小
func (d *Map3S[T1, T2, T3, V]) Size() int {
	return len(d.data)
}

// 设置
func (d *Map3S[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	d.mutex.Lock()
	d.data[Triple[T1, T2, T3]{t1, t2, t3}] = value
	d.mutex.Unlock()
}

// 读取
func (d *Map3S[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	d.mutex.RLock()
	value, ok := d.data[Triple[T1, T2, T3]{t1, t2, t3}]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Map3S[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
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

func (d *Map3S[T1, T2, T3, V]) Walk(f func(V) bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, item := range d.data {
		if !f(item) {
			return
		}
	}
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
	d.mutex.Lock()
	d.data[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}] = value
	d.mutex.Unlock()
}

// 读取
func (d *Map4S[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	d.mutex.RLock()
	value, ok := d.data[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}]
	d.mutex.RUnlock()
	return value, ok
}

// 删除
func (d *Map4S[T1, T2, T3, T4, V]) Del(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	key := Quad[T1, T2, T3, T4]{t1, t2, t3, t4}
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

func (d *Map4S[T1, T2, T3, T4, V]) Walk(f func(V) bool) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, item := range d.data {
		if !f(item) {
			return
		}
	}
}
