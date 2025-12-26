package structure

type Map2[T1 comparable, T2 comparable, V any] map[Pair[T1, T2]]V

func NewMap2[T1 comparable, T2 comparable, V any]() Map2[T1, T2, V] {
	return make(map[Pair[T1, T2]]V)
}

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

func (d Map2[T1, T2, V]) Copy(t1 T1, t2 T2, f func(V) V) (V, bool) {
	value, ok := d[Pair[T1, T2]{t1, t2}]
	if ok {
		value = f(value)
	}
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
