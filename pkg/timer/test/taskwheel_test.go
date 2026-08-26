package timer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/timer"
	"github.com/hechh/framework/pkg/timer/adapter/taskwheel"
)

// Test_TaskWheel_Register 使用 taskwheel 库的定时器注册测试
// 与 Test_Timer_Register 逻辑一致，仅替换底层实现做对比。
//
// 测试结论（100000 个任务，每个 TTL=1s，重复 5 次，等待 6.9s）：
//   - lockfree_timer（core/fun/network/adapter/websocket/）: ~550k 次触发
//   - taskwheel（github.com/ankur-anand/taskwheel）:          ~600k 次触发
//
// taskwheel 在此场景下触发数略高，主要得益于其 1ms 粒度 tick（vs lockfree 的 32ms），
// 使得时间精度更高、任务到期时更少错过 tick 窗口。
func Test_TaskWheel_Register(t *testing.T) {
	cfg := &timer.Config{
		Size:               4,
		MinPeriodBitNumber: 0,
	}
	ot := timer.NewTimer(taskwheel.NewTimer)
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
	t.Log("------>taskwheel count:", final)
	if final < 3500000 {
		t.Fatalf("触发次数不足: 期望至少 3500000，实际 %d", final)
	}
}
