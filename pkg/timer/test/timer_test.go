package timer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/timer"
	"github.com/hechh/framework/pkg/timer/adapter/lockfree_timer"
)

func Test_Timer_Register(t *testing.T) {
	cfg := &timer.Config{
		Size:               6,
		MinPeriodBitNumber: 6,
	}
	ot := timer.NewTimer(lockfree_timer.NewTimer)
	if err := ot.Init(cfg); err != nil {
		t.Fatalf("timer 初始化失败， error=%v", err)
	}

	timer.SetObject(ot)
	defer timer.Close()

	var count int32
	for range 1000000 {
		id := uint64(1)
		task := timer.NewTask(&id, time.Second, 5, func() {
			atomic.AddInt32(&count, 1)
		})
		err := ot.Register(task)
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
	}
	time.Sleep(5900 * time.Millisecond)
	final := atomic.LoadInt32(&count)
	t.Log("------>", final)
	if final < 4500000 {
		t.Fatalf("触发次数不足: 期望至少 450000，实际 %d", final)
	}
}
