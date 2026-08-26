package gc

var object *Gc

func SetObject(oj *Gc) {
	object = oj
}

func Destroy(f func()) {
	if object != nil {
		object.Push(f)
	}
}
