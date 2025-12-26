package structure

type Map1[T1 comparable, V any] map[T1]V

func (d Map1[T1, V]) Size() int {
	return len(d)
}

func (d Map1[T1, V]) Put(t1 T1, value V) {
	d[t1] = value
}

func (d Map1[T1, V]) Get(t1 T1) (V, bool) {
	value, ok := d[t1]
	return value, ok
}

func (d Map1[T1, V]) Copy(t1 T1, f func(V) V) (V, bool) {
	value, ok := d[t1]
	if ok {
		value = f(value)
	}
	return value, ok
}

func (d Map1[T1, V]) Del(t1 T1) (V, bool) {
	value, ok := d[t1]
	delete(d, t1)
	return value, ok
}
