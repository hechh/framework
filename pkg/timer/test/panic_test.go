package timer

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/timer"
	"github.com/hechh/framework/pkg/timer/adapter/lockfree_timer"
	"github.com/hechh/framework/pkg/timer/adapter/taskwheel"
)

// ============================================================
// 定时任务回调 panic 的隔离验证：单个任务 panic 不能崩进程，
// 也不能让定时器本身停摆。
// ============================================================

// Test_TaskWheel_TaskPanicIsolated 验证 taskwheel 适配器下回调 panic 的隔离。
//
// 修复前：回调由 taskwheel 的 tick goroutine 裸调（库内无 recover），
// 业务回调 panic 直接使整个进程 fatal。
func Test_TaskWheel_TaskPanicIsolated(t *testing.T) {
	ot := timer.NewTimer(taskwheel.NewTimer)
	if err := ot.Init(&timer.Config{Size: 4, MinPeriodBitNumber: 0}); err != nil {
		t.Fatalf("timer 初始化失败， error=%v", err)
	}
	defer ot.Close()

	var boom, alive int32
	boomID, aliveID := uint64(1), uint64(2)
	// 首次执行 panic，之后正常返回：既能验证 panic 被隔离，也不至于刷屏
	panicTask := timer.NewTask(&boomID, 20*time.Millisecond, -1, func() {
		if atomic.AddInt32(&boom, 1) == 1 {
			panic("测试用 panic：业务回调崩溃")
		}
	})
	normalTask := timer.NewTask(&aliveID, 20*time.Millisecond, -1, func() {
		atomic.AddInt32(&alive, 1)
	})
	if err := ot.Register(panicTask); err != nil {
		t.Fatalf("Register panicTask failed: %v", err)
	}
	if err := ot.Register(normalTask); err != nil {
		t.Fatalf("Register normalTask failed: %v", err)
	}

	time.Sleep(400 * time.Millisecond)

	if atomic.LoadInt32(&boom) == 0 {
		t.Fatalf("panic 任务未执行")
	}
	// panic 发生在第一个 tick 批内，同批及后续 tick 必须继续调度
	if got := atomic.LoadInt32(&alive); got < 3 {
		t.Fatalf("panic 后定时器停摆：正常任务仅触发 %d 次", got)
	}
}

// Test_Lockfree_TaskPanicKeepsWorkers 验证 lockfree 适配器的 worker 不因回调 panic 退出。
//
// 修复前：consume 中的 panic 逃逸到 safe.SafeGo 后该 worker goroutine 永久退出，
// 每个 panic 永久少一个 worker；Size 个 worker 被打光后全进程定时器静默停摆
// （且 panic 的任务已从轮中弹出，不会重入轮）。
func Test_Lockfree_TaskPanicKeepsWorkers(t *testing.T) {
	const size = 2
	ot := timer.NewTimer(lockfree_timer.NewTimer)
	if err := ot.Init(&timer.Config{Size: size, MinPeriodBitNumber: 5}); err != nil {
		t.Fatalf("timer 初始化失败， error=%v", err)
	}
	defer ot.Close()

	// 每个任务 panic 一次：Size 个 panic 足以打光全部 worker
	var boom int32
	for i := 0; i < size; i++ {
		id := uint64(i + 1)
		task := timer.NewTask(&id, 32*time.Millisecond, -1, func() {
			if atomic.AddInt32(&boom, 1) <= size {
				panic("测试用 panic：业务回调崩溃")
			}
		})
		if err := ot.Register(task); err != nil {
			t.Fatalf("Register panicTask failed: %v", err)
		}
	}
	// 等待全部 panic 触发（worker 若已死光，此后不会再有任务被执行）
	time.Sleep(300 * time.Millisecond)

	var alive int32
	aliveID := uint64(100)
	normalTask := timer.NewTask(&aliveID, 32*time.Millisecond, -1, func() {
		atomic.AddInt32(&alive, 1)
	})
	if err := ot.Register(normalTask); err != nil {
		t.Fatalf("Register normalTask failed: %v", err)
	}
	time.Sleep(400 * time.Millisecond)

	if got := atomic.LoadInt32(&alive); got == 0 {
		t.Fatalf("worker 被 panic 打光后定时器停摆：正常任务一次都未触发")
	}
}
