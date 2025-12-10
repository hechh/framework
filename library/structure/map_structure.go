package structure

type Map2[T1 comparable, T2 comparable, V any] map[Pair[T1, T2]]V

// 大小
func (d Map2[T1, T2, V]) Size() int {
	return len(d)
}

// 设置
func (d Map2[T1, T2, V]) Set(t1 T1, t2 T2, value V) {
	d[Pair[T1, T2]{t1, t2}] = value
}

// 读取
func (d Map2[T1, T2, V]) Get(t1 T1, t2 T2) (V, bool) {
	value, ok := d[Pair[T1, T2]{t1, t2}]
	return value, ok
}

// 删除
func (d Map2[T1, T2, V]) Del(t1 T1, t2 T2) (V, bool) {
	key := Pair[T1, T2]{t1, t2}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Map3[T1 comparable, T2 comparable, T3 comparable, V any] map[Triple[T1, T2, T3]]V

// 大小
func (d Map3[T1, T2, T3, V]) Size() int {
	return len(d)
}

// 设置
func (d Map3[T1, T2, T3, V]) Set(t1 T1, t2 T2, t3 T3, value V) {
	d[Triple[T1, T2, T3]{t1, t2, t3}] = value
}

// 读取
func (d Map3[T1, T2, T3, V]) Get(t1 T1, t2 T2, t3 T3) (V, bool) {
	value, ok := d[Triple[T1, T2, T3]{t1, t2, t3}]
	return value, ok
}

// 删除
func (d Map3[T1, T2, T3, V]) Del(t1 T1, t2 T2, t3 T3) (V, bool) {
	key := Triple[T1, T2, T3]{t1, t2, t3}
	value, ok := d[key]
	if ok {
		delete(d, key)
	}
	return value, ok
}

type Map4[T1 comparable, T2 comparable, T3 comparable, T4 comparable, V any] map[Quad[T1, T2, T3, T4]]V

// 大小
func (d Map4[T1, T2, T3, T4, V]) Size() int {
	return len(d)
}

// 设置
func (d Map4[T1, T2, T3, T4, V]) Set(t1 T1, t2 T2, t3 T3, t4 T4, value V) {
	d[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}] = value
}

// 读取
func (d Map4[T1, T2, T3, T4, V]) Get(t1 T1, t2 T2, t3 T3, t4 T4) (V, bool) {
	value, ok := d[Quad[T1, T2, T3, T4]{t1, t2, t3, t4}]
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
