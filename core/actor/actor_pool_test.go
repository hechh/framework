package actor

import (
	"testing"
	"time"

	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/queue"
	"github.com/hechh/framework/packet"
)

// poolTestActor 受控桩：仅用于 ActorPool.Register 的参数校验
type poolTestActor struct{}

func (*poolTestActor) GetName() string                                  { return "poolTestActor" }
func (*poolTestActor) GetId() uint64                                    { return 0 }
func (*poolTestActor) Start() error                                     { return nil }
func (*poolTestActor) Stop()                                            {}
func (*poolTestActor) Register(IActor, define.ICache, ...queue.Option)  {}
func (*poolTestActor) RegisterTimer(string, time.Duration, int32) error { return nil }
func (*poolTestActor) SendMsg(*packet.Head, ...any) error               { return nil }
func (*poolTestActor) Send(*packet.Head, []byte) error                  { return nil }

// TestActorPoolRegister_AppliesOpts 验证 opts 被传入协程池（size 与 name 生效）。
//
// 修复前 opts 被丢弃：size 默认 0 → taskCh 无缓冲且启动 0 个 worker，
// 首个任务在 handle() 的 taskCh <- f 处永久阻塞，并连带卡死 gc 单协程的全进程 actor 清理。
func TestActorPoolRegister_AppliesOpts(t *testing.T) {
	pool := &ActorPool{}
	pool.Register(&poolTestActor{}, nil, queue.WithSize(4))

	if got := pool.msgs.GetSize(); got != 4 {
		t.Fatalf("Register 未把 opts 传给协程池: size=%d, want=4", got)
	}
	if got := pool.msgs.GetName(); got != "poolTestActor" {
		t.Fatalf("协程池名称未设置: name=%q", got)
	}
}

// TestActorPoolRegister_RejectsZeroSize 验证未设置协程池大小时初始化直接失败（尽早暴露配置错误）
func TestActorPoolRegister_RejectsZeroSize(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("未设置协程池大小时应初始化失败（size=0 会导致任务永久阻塞）")
		}
	}()

	pool := &ActorPool{}
	pool.Register(&poolTestActor{}, nil)
}
