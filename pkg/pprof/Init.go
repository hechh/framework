package pprof

var object *Pprof

func SetObject(oj *Pprof) {
	object = oj
}

func Close() {
	if object != nil {
		object.Close()
	}
}
