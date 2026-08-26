package timer

import (
	"fmt"
)

var (
	object *Timer
)

func SetObject(oj *Timer) {
	object = oj
}

func GetObject() *Timer {
	return object
}

// 注册定时任务
func Register(task ITask) error {
	if object != nil {
		return object.Register(task)
	}
	return fmt.Errorf("timer not initialized")
}

func Close() {
	if object != nil {
		object.Close()
	}
}
