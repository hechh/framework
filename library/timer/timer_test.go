package timer

import (
	"fmt"
	"framework/library/safe"
	"testing"
	"time"
)

func TestTimer(t *testing.T) {
	safe.SetExcept(t.Logf)
	timer := NewTimer(4, 5)
	taskId := uint64(123)
	for i := 0; i < 2; i++ {
		err := timer.Register(&taskId, 1*time.Second, -1, func() {
			fmt.Println("-->", i, time.Now().Unix())
		})
		if err != nil {
			t.Log("Register failed", err)
			return
		}
	}
	time.Sleep(4 * time.Second)
	//select {}
}
