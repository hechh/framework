package redispool

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/hechh/framework/core/context"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
)

// fakeClient 连接池测试替身：内嵌 IClient 接口，只实现用例真正调用的方法。
type fakeClient struct {
	IClient
	name    string
	closed  *int32
	msetErr error
}

func (c *fakeClient) DbName() string { return c.name }

func (c *fakeClient) Close() error {
	if c.closed != nil {
		atomic.AddInt32(c.closed, 1)
	}
	return nil
}

func (c *fakeClient) MSet(args ...any) error              { return c.msetErr }
func (c *fakeClient) HMSet(key string, vals ...any) error { return c.msetErr }

// newPoolWith 构造使用 fakeClient 的连接池；dbname 等于 failName 的连接写入必失败，
// closed 记录累计关闭次数（用于校验失败路径不泄漏连接）。
func newPoolWith(closed *int32, failName string) *RedisPool {
	created := new(int32)
	return newPoolCounting(closed, created, failName)
}

// newPoolCounting 同上，额外用 created 记录创建的连接数（校验"创建即关闭"）
func newPoolCounting(closed, created *int32, failName string) *RedisPool {
	return NewRedisPool(func(cfg *Config) (IClient, error) {
		atomic.AddInt32(created, 1)
		cli := &fakeClient{name: cfg.DbName, closed: closed}
		if cfg.DbName == failName {
			cli.msetErr = errors.New("模拟Redis写入失败")
		}
		return cli, nil
	})
}

// memCache 常驻缓存测试替身
type memCache struct {
	items map[string]define.IValue
}

func (m *memCache) Has(key string) bool                    { _, ok := m.items[key]; return ok }
func (m *memCache) SetCache(key string, val define.IValue) { m.items[key] = val }
func (m *memCache) GetCache(key string) define.IValue      { return m.items[key] }
func (m *memCache) GetAllCache() map[string]define.IValue  { return m.items }
func (m *memCache) Refresh(except ...string)               {}

// TestInit_RejectsDuplicateDbName 验证同名连接被拒绝而不是静默覆盖。
//
// dbname 同时是连接池键、一致性哈希节点 id、批量写分组键与迁移的"同库"判断依据；
// 重名会让后创建的连接静默覆盖先创建的（被覆盖的连接永不关闭 → 连接泄漏，
// 且按名称取客户端会拿到错误的库），迁移也会误判"同库"而跳过迁移。
// 这里用同名 globals：它不参与哈希环，修复前连报错都不会有。
func TestInit_RejectsDuplicateDbName(t *testing.T) {
	var closed, created int32
	pool := newPoolCounting(&closed, &created, "")

	err := pool.Init([]*Config{{DbName: "global"}, {DbName: "global"}}, nil)
	if err == nil {
		t.Fatalf("同名 dbname 必须报错：否则连接被静默覆盖（旧连接泄漏、数据落错库）")
	}
	// 失败路径必须关闭全部已创建的连接
	if got, want := atomic.LoadInt32(&closed), atomic.LoadInt32(&created); got != want || want == 0 {
		t.Fatalf("Init 失败后必须关闭已创建的连接: created=%d closed=%d", want, got)
	}
}

// TestInit_ClosesClientsOnAddNodeError 验证哈希环注册失败时连接不泄漏。
//
// 修复前 AddNode 失败分支直接 return err（没有 d.Close()），已创建的连接全部泄漏；
// 同名分片会走到该分支（旧连接被池覆盖后既不关闭、也无法再通过池访问）。
func TestInit_ClosesClientsOnAddNodeError(t *testing.T) {
	var closed, created int32
	pool := newPoolCounting(&closed, &created, "")

	err := pool.Init([]*Config{{DbName: "global"}}, []*Config{{DbName: "player_1"}, {DbName: "player_1"}})
	if err == nil {
		t.Fatalf("同名分片必须报错：否则连接被静默覆盖且数据会落错库")
	}
	if got, want := atomic.LoadInt32(&closed), atomic.LoadInt32(&created); got != want {
		t.Fatalf("哈希环注册失败后必须关闭全部已创建的连接: created=%d closed=%d", want, got)
	}
}

// TestInit_RejectsEmptyConfig 验证 globals 与 shards 全空时直接报错，避免带病启动
func TestInit_RejectsEmptyConfig(t *testing.T) {
	var closed int32
	if err := newPoolWith(&closed, "").Init(nil, nil); err == nil {
		t.Fatalf("globals 与 shards 全空必须报错")
	}
}

// TestInit_FreezesHashRing 验证 Init 完成后哈希环已 Freeze。
//
// StaticHash 的无锁读以"构建完成后不再写入"为安全前提（其文档明确要求调用 Freeze）；
// 未 Freeze 时若运行期再 AddNode，读路径会与 h.hashRing = merged 并发（切片头 3 个字
// 非原子写）→ 撕裂 slice 头崩溃或路由错乱串号。
func TestInit_FreezesHashRing(t *testing.T) {
	var closed int32
	pool := newPoolWith(&closed, "")
	if err := pool.Init([]*Config{{DbName: "global"}}, []*Config{{DbName: "player_1"}}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer pool.Close()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Init 后必须 Freeze 哈希环：运行期 AddNode 会与无锁读并发改写 hashRing")
		}
	}()
	_ = pool.virtuals.AddNode("player_2", &fakeClient{name: "player_2"})
}

// TestSaveByCtx_PartialFailureKeepsCacheConsistent 验证部分分组写入失败时的缓存一致性。
//
// 修复前：Save 按连接分组独立写、失败只记日志继续，而 SaveByCtx 在 err 时提前 return
// 不 Refresh → 常驻缓存仍是旧对象，下一个请求的 Save 会把旧值写回已写成功的分组，
// 玩家已收到成功响应却丢奖励（"已领取标记"回滚还会导致二次领取）。
func TestSaveByCtx_PartialFailureKeepsCacheConsistent(t *testing.T) {
	var closed int32
	pool := newPoolWith(&closed, "global") // 全局库写入失败
	if err := pool.Init([]*Config{{DbName: "global"}}, []*Config{{DbName: "player_1"}}); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer pool.Close()

	shardVal := NewValue(pool.Get("player_1"), &packet.Head{}, STRING, "k_shard")
	globalVal := NewValue(pool.Get("global"), &packet.Head{}, STRING, "k_global")
	shardVal.Change()
	globalVal.Change()

	// 常驻缓存中保存的是本次修改前的旧对象（请求开始时按 key 克隆出来放进临时缓存）
	oldShard := NewValue(pool.Get("player_1"), &packet.Head{}, STRING, "k_shard")
	oldGlobal := NewValue(pool.Get("global"), &packet.Head{}, STRING, "k_global")

	cache := &memCache{items: map[string]define.IValue{
		"k_shard":  oldShard,
		"k_global": oldGlobal,
	}}
	ctx := context.NewContext(&packet.Head{}, cache)
	ctx.SetCache("k_shard", shardVal)
	ctx.SetCache("k_global", globalVal)

	if err := SaveByCtx(ctx); err == nil {
		t.Fatalf("全局库写入失败必须返回错误")
	}
	// 写成功的分组必须提交到常驻缓存：否则下次 Save 会用旧缓存把新值覆盖回 Redis
	if cache.GetCache("k_shard") != shardVal {
		t.Fatalf("写成功的分组必须提交到常驻缓存，否则下次 Save 会把旧值写回 Redis")
	}
	// 写失败的分组不得提交：常驻缓存必须与 Redis（旧值）保持一致
	if cache.GetCache("k_global") != oldGlobal {
		t.Fatalf("写失败的分组不能提交到常驻缓存，否则与 Redis 永久不一致")
	}
	if !globalVal.IsChanged() {
		t.Fatalf("写失败的分组必须保留脏标记，供下次 Save 重试")
	}
}
