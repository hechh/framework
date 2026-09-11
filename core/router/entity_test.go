package router

import (
	"testing"
	"time"

	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/pkg/gc"
)

// TestNewEntity_BindsParent 验证 NewEntity 绑定了 parent。
//
// 缺此赋值时 Entity.Call() 里的 parent.Remove 永不执行：时间轮触发一次后实体不再重注册，
// 却仍留在 shard.data，路由表随时间无界增长（"过期淘汰"设计完全失效）。
func TestNewEntity_BindsParent(t *testing.T) {
	r := NewRouter()
	e := NewEntity(0, 1, r)
	if e.parent != r {
		t.Fatal("NewEntity 未绑定 parent（过期淘汰不会从路由表移除，路由表将无界增长）")
	}
}

// TestEntityCall_RemovesFromRouter 验证实体过期（times 归零）后从路由表移除。
func TestEntityCall_RemovesFromRouter(t *testing.T) {
	// 过期回调经 gc 异步执行，测试内初始化 gc 组件
	g := &gc.Gc{}
	if err := g.Init(); err != nil {
		t.Fatal(err)
	}
	gc.SetObject(g)
	defer func() {
		g.Close()
		gc.SetObject(nil)
	}()

	const (
		idType uint32 = 0
		id     uint64 = 7
	)

	r := NewRouter()
	e := NewEntity(idType, id, r)
	// 直接写入分片：GetOrNew 会注册时间轮（依赖 timer 组件），此处聚焦"过期淘汰 → 移除"链路
	shard := r.shards[id%256]
	shard.mutex.Lock()
	shard.data[tplutil.T2(idType, id)] = e
	shard.mutex.Unlock()

	e.Call() // 时间轮到期的回调：times 归零 → gc 异步执行 parent.Remove

	deadline := time.Now().Add(2 * time.Second)
	for r.Get(idType, id) != nil {
		if time.Now().After(deadline) {
			t.Fatal("实体过期后仍留在路由表（parent 未绑定或 Remove 未执行）")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
