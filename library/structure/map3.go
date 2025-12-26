package structure

type Map3[T1 comparable, T2 comparable, T3 comparable, V any] map[Triple[T1, T2, T3]]V

func NewMap3[T1 comparable, T2 comparable, T3 comparable, V any]() Map3[T1, T2, T3, V] {
	return make(map[Triple[T1, T2, T3]]V)
}

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

func (d Map3[T1, T2, T3, V]) Copy(t1 T1, t2 T2, t3 T3, f func(V) V) (V, bool) {
	value, ok := d[Triple[T1, T2, T3]{t1, t2, t3}]
	if ok {
		value = f(value)
	}
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
