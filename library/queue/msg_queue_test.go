package queue

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// testTask 测试任务：Do 执行 fn 并返回 true（表示"处理过任务"，用于刷新 updateTime）
type testTask struct {
	fn func()
}

func (d *testTask) Do() bool {
	if d.fn != nil {
		d.fn()
	}
	return true
}

// panicRecorder 记录 SafeGo(except, ...) 捕获到的 panic（run/worker 内 panic 会被 recover 并回调 except）
type panicRecorder struct {
	mu   sync.Mutex
	msgs []string
}

func (d *panicRecorder) except(format string, args ...any) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.msgs = append(d.msgs, format)
}

func (d *panicRecorder) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.msgs)
}

// unlockProbe 记录解锁调用：ID 与"已调用"标志
type unlockProbe struct {
	mu     sync.Mutex
	called bool
	id     uint64
}

func (d *unlockProbe) unlock(id uint64) error {
	d.mu.Lock()
	d.called = true
	d.id = id
	d.mu.Unlock()
	return nil
}

func (d *unlockProbe) get() (bool, uint64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.called, d.id
}

// 解锁耗时：大于调度抖动，用于断言"Wait() 是否等到了解锁完成"
const unlockCost = 200 * time.Millisecond

// ---------- C15：Wait() 必须在解锁完成后才返回 ----------

// TestMsgQueueWaitWaitsForUnlock 验证 Wait() 返回时 OnUnlock 已完成，且解锁用的是启动时的 actor id。
//
// 回归背景：run 的 defer 曾以 wg.Done() 作为首条语句，而 Wait() 在 wg 归零后立刻 SetId(0)，
// 此时 run 仍要执行 handle()/OnUnlock()。解锁晚于 SetId(0) 就会解 actor_locker:0，
// 真实的 actor_locker:<uid> 无人释放（TTL 最长 12 分钟，玩家无法重登/被接管）。
//
// 断言方式：让 unlocker 睡 unlockCost。旧实现下 Wait() 在解锁完成前就返回（耗时≈0），
// 新实现必须等到解锁结束（耗时≥unlockCost），因此该断言是确定性的负向对照。
func TestMsgQueueWaitWaitsForUnlock(t *testing.T) {
	const actorId uint64 = 123456

	probe := &unlockProbe{}
	q := NewMsgQueue[*testTask](
		WithId(actorId),
		WithLocker(0,
			func(id uint64, expire time.Duration) error { return nil },
			func(id uint64) error {
				time.Sleep(unlockCost)
				return probe.unlock(id)
			},
		),
	)

	if err := q.Start(nil); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	q.Stop()

	begin := time.Now()
	q.Wait()
	elapsed := time.Since(begin)

	called, gotId := probe.get()
	if !called {
		t.Errorf("Wait() 返回时 OnUnlock 尚未执行（收尾未完成就宣告关闭）")
	}
	if gotId != actorId {
		t.Errorf("解锁使用的 id=%d, 期望 %d（说明解锁到了 actor_locker:0，真实锁泄漏）", gotId, actorId)
	}
	if elapsed < unlockCost {
		t.Errorf("Wait() 仅耗时 %v（< %v），说明没有等待 OnUnlock 完成", elapsed, unlockCost)
	}
	if !q.IsClosed() {
		t.Errorf("Wait() 后状态=%d, 期望 STOPPED", q.GetStatus())
	}
	if q.GetId() != 0 {
		t.Errorf("Wait() 后 id=%d, 期望 0", q.GetId())
	}
}

// TestMsgQueueWaitDrainsTasks 验证收尾阶段仍会处理队列中剩余的任务
func TestMsgQueueWaitDrainsTasks(t *testing.T) {
	const taskCount = 50

	var done atomic.Int32
	q := NewMsgQueue[*testTask](WithId(1))
	if err := q.Start(nil); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}

	for range taskCount {
		q.Push(&testTask{fn: func() { done.Add(1) }})
	}
	// 等 run 把任务消费完再关闭，避免依赖内部批处理细节
	deadline := time.Now().Add(2 * time.Second)
	for done.Load() < taskCount && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	q.Stop()
	q.Wait()

	if got := done.Load(); got != taskCount {
		t.Errorf("任务执行数=%d, 期望 %d", got, taskCount)
	}
}

// ---------- C16：协程池关闭必须等到 worker 退出，且不能 send on closed channel ----------

// TestMsgQueuePoolWaitWaitsForUnlockAndWorkers 验证协程池关闭时：
//  1. Wait() 等到 run 收尾（含解锁）完成 —— 与 C15 同一根因（w1.Done() 为 defer 首条语句）
//  2. 关闭时队列仍有任务，也不出现 send on closed channel（panic 会跳过解锁，锁泄漏到 TTL）
//  3. Wait() 返回后不再有 worker 在跑（w2 由"退出时计数"改为 defer 后才成立）
func TestMsgQueuePoolWaitWaitsForUnlockAndWorkers(t *testing.T) {
	const (
		actorId   uint64 = 654321
		taskCount        = 300
	)

	var (
		done   atomic.Int32
		probe  = &unlockProbe{}
		panics = &panicRecorder{}
	)

	q := NewMsgQueuePool[*testTask](
		WithId(actorId),
		WithSize(2),
		WithLocker(0,
			func(id uint64, expire time.Duration) error { return nil },
			func(id uint64) error {
				time.Sleep(unlockCost)
				return probe.unlock(id)
			},
		),
	)

	if err := q.Start(panics.except); err != nil {
		t.Fatalf("Start 失败: %v", err)
	}
	for range taskCount {
		q.Push(&testTask{fn: func() { done.Add(1) }})
	}

	// 立即关闭：制造"关闭时队列仍有任务"路径（正是旧实现 panic 的触发条件）
	q.Stop()

	begin := time.Now()
	q.Wait()
	elapsed := time.Since(begin)

	if got := panics.count(); got != 0 {
		t.Errorf("收尾阶段发生 %d 次 panic（send on closed channel 会跳过解锁）", got)
	}

	called, gotId := probe.get()
	if !called {
		t.Errorf("Wait() 返回时 OnUnlock 尚未执行")
	}
	if gotId != actorId {
		t.Errorf("解锁使用的 id=%d, 期望 %d", gotId, actorId)
	}
	if elapsed < unlockCost {
		t.Errorf("Wait() 仅耗时 %v（< %v），说明没有等待收尾完成", elapsed, unlockCost)
	}

	// Wait() 返回后 worker 必须已全部退出：任务计数不再增长
	afterWait := done.Load()
	time.Sleep(100 * time.Millisecond)
	if grew := done.Load(); grew != afterWait {
		t.Errorf("Wait() 返回后仍有 worker 在消费任务: %d → %d（w2 未真正等待 worker 退出）", afterWait, grew)
	}
	if afterWait == 0 {
		t.Errorf("关闭时队列中的任务一个都没执行（剩余任务被静默丢弃）")
	}
	if !q.IsClosed() {
		t.Errorf("Wait() 后状态=%d, 期望 STOPPED", q.GetStatus())
	}
}

// TestMsgQueuePoolStartStopRace 反复启停，配合 -race 捕捉启停路径的数据竞争
func TestMsgQueuePoolStartStopRace(t *testing.T) {
	for i := range 50 {
		var done atomic.Int32
		q := NewMsgQueuePool[*testTask](WithId(uint64(i+1)), WithSize(2))
		if err := q.Start(nil); err != nil {
			t.Fatalf("第 %d 轮 Start 失败: %v", i, err)
		}
		for range 10 {
			q.Push(&testTask{fn: func() { done.Add(1) }})
		}
		q.Stop()
		q.Wait()
		if !q.IsClosed() {
			t.Fatalf("第 %d 轮 Wait 后状态=%d, 期望 STOPPED", i, q.GetStatus())
		}
	}
}

// ---------- C17：抢锁失败必须回滚状态，重复 Start 必须显式报错 ----------

// TestMsgQueueStartLockFailedRollsBack 验证抢锁失败后状态回滚到 STOPPED（可重试），
// 且运行中重复 Start 返回显式错误（旧实现两处都表现为"静默返回 nil"）。
func TestMsgQueueStartLockFailedRollsBack(t *testing.T) {
	var lockCalls atomic.Int32
	lockErr := errors.New("锁被其他节点持有")

	q := NewMsgQueue[*testTask](
		WithId(1),
		WithLocker(10,
			func(id uint64, expire time.Duration) error {
				if lockCalls.Add(1) == 1 {
					return lockErr
				}
				return nil
			},
			func(id uint64) error { return nil },
		),
	)

	// ① 首次抢锁失败：返回错误，且状态回滚（旧实现永久停留在 STARTING）
	if err := q.Start(nil); !errors.Is(err, lockErr) {
		t.Fatalf("抢锁失败应返回原错误, 实际 err=%v", err)
	}
	if !q.IsClosed() {
		t.Errorf("抢锁失败后状态=%d, 期望 STOPPED（未回滚会让后续 Start 静默失效、Stop/Wait 永久不可用）", q.GetStatus())
	}

	// ② 回滚后可重试成功
	if err := q.Start(nil); err != nil {
		t.Fatalf("状态回滚后重试仍失败: %v", err)
	}
	if !q.IsRunning() {
		t.Errorf("重试成功后状态=%d, 期望 RUNNING", q.GetStatus())
	}

	// ③ 运行中重复 Start：必须显式报错而不是 nil
	if err := q.Start(nil); err == nil {
		t.Errorf("已启动的队列重复 Start 应返回显式错误, 实际 nil（调用方会误判为启动成功）")
	}

	// ④ 收尾仍正常
	q.Stop()
	q.Wait()
	if !q.IsClosed() {
		t.Errorf("Wait() 后状态=%d, 期望 STOPPED", q.GetStatus())
	}
}

// TestMsgQueuePoolStartLockFailedRollsBack 协程池同构：抢锁失败回滚 + 重复 Start 报错
func TestMsgQueuePoolStartLockFailedRollsBack(t *testing.T) {
	var lockCalls atomic.Int32
	lockErr := errors.New("锁被其他节点持有")

	q := NewMsgQueuePool[*testTask](
		WithId(1),
		WithSize(2),
		WithLocker(10,
			func(id uint64, expire time.Duration) error {
				if lockCalls.Add(1) == 1 {
					return lockErr
				}
				return nil
			},
			func(id uint64) error { return nil },
		),
	)

	if err := q.Start(nil); !errors.Is(err, lockErr) {
		t.Fatalf("抢锁失败应返回原错误, 实际 err=%v", err)
	}
	if !q.IsClosed() {
		t.Errorf("抢锁失败后状态=%d, 期望 STOPPED", q.GetStatus())
	}

	if err := q.Start(nil); err != nil {
		t.Fatalf("状态回滚后重试仍失败: %v", err)
	}
	if err := q.Start(nil); err == nil {
		t.Errorf("已启动的协程池重复 Start 应返回显式错误, 实际 nil")
	}

	q.Stop()
	q.Wait()
	if !q.IsClosed() {
		t.Errorf("Wait() 后状态=%d, 期望 STOPPED", q.GetStatus())
	}
}
