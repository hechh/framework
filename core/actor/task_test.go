package actor

import (
	"testing"

	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
)

// panicHandler 受控桩：Call 必定 panic，用于验证 Task.Do 的 recover 与任务释放
type panicHandler struct{}

func (panicHandler) GetActorFuncName() string { return "Test.Panic" }
func (panicHandler) GetActorFunc() uint32     { return 1 }
func (panicHandler) GetMask() uint32          { return define.CMD_FLAG }
func (panicHandler) Call(any, define.IContext, ...any) error {
	panic("boom")
}

// TestTaskDo_PanicRecoveredAndTaskReleased 验证任务执行 panic 时：
//  1. panic 不逃出 Do()——修复前 defer 内先 Release（IContext 置 nil 并归还对象池）再 recover，
//     recover 分支访问已释放字段会二次 panic，新 panic 逃出 Do() 会终止队列协程、
//     连带销毁整个 Actor（与"单任务崩溃不拖垮 Actor"的设计相悖）；
//  2. 任务对象仍被正确释放，可安全归还对象池。
func TestTaskDo_PanicRecoveredAndTaskReleased(t *testing.T) {
	head := packet.GetHead()
	// head 所有权转移给任务，由 Task.Release 内的 Context.Destroy 归还，测试不重复 PutHead

	task := NewTask(nil, nil, panicHandler{}, head, []any{})

	defer func() {
		if e := recover(); e != nil {
			t.Fatalf("panic 逃出 Task.Do(): %v", e)
		}
	}()
	task.Do()

	if task.IContext != nil {
		t.Fatal("Task.Do 结束后任务未被释放（IContext 应被置 nil）")
	}
}
