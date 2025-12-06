package async

import (
	"runtime/debug"
)

var (
	except func(string, ...any)
)

func SetExcept(e func(string, ...any)) {
	except = e
}

func Go(f func()) {
	go func() {
		Recover(f)
	}()
}

func Recover(f func()) {
	defer func() {
		if err := recover(); err != nil && except != nil {
			except("%v stack: %v", err, string(debug.Stack()))
		}
	}()
	f()
}
