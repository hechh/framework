package consistent

import (
	"fmt"
	"sync/atomic"
	"testing"
)

// benchSinkTotal 各 P 结束后把局部累加结果一次性合入（防编译器消除；不引入 per-query 原子开销）
var benchSinkTotal atomic.Uint64

// ---------------------------------------------------------------
// 一致性哈希 性能基准测试（benchmark）
//
// 读路径成本拆解：
//  1. RWMutex.RLock（并发安全，锁竞争随 goroutine 数上升）
//  2. key 哈希：GetNode(string) 走 FNV-64a + mix64（有接口逃逸分配）；
//     GetNodeByHash 仅 mix64（无分配）
//  3. 二分查找 sort.Search：O(log2(nodes×virtuals))
//
// 运行：
//   go test -bench . -benchmem -run '^$' ./library/consistent/
//
// 可选参数：
//   -benchtime=1s   每基准运行时长
//   -cpu=1,8,16     P 数（测锁竞争 / 并发扩展性）
// ---------------------------------------------------------------

// 覆盖典型集群规模：小 / 中 / 大 / 超大
var benchNodeCounts = []int{10, 100, 1000, 2000}

const (
	benchVirtuals   = 150 // 默认虚拟节点数（与生产一致）
	benchStringPool = 100000
	benchUintPool   = 1 << 20
)

// benchBuildHash 构建指定规模的一致性哈希（节点 id 从 1 开始）
func benchBuildHash(b *testing.B, nodeCount, virtuals int) *Hash[uint32, *testNode] {
	b.Helper()
	h := NewHash[uint32, *testNode](virtuals)
	for _, n := range newTestNodesFrom(1, nodeCount) {
		if err := h.AddNode(n.id, n); err != nil {
			b.Fatalf("AddNode(%d): %v", n.id, err)
		}
	}
	return h
}

// benchStringKeys 预生成 key 池（避免把 fmt.Sprintf 的开销计入基准）
func benchStringKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("player-%08d", i)
	}
	return keys
}

// benchLongKeys 生成长 key（128 字节），放大 FNV 哈希计算成本
func benchLongKeys(n int) []string {
	keys := make([]string, n)
	for i := range keys {
		keys[i] = fmt.Sprintf("player-%08d-abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz0123456789", i)
	}
	return keys
}

// benchUintKeys 预生成伪随机 uint64 key 池（贴近真实递增 ID 散列后的形态）
func benchUintKeys(n uint64) []uint64 {
	keys := make([]uint64, n)
	x := uint64(0x9E3779B97F4A7C15) // 黄金分割种子
	for i := range keys {
		x = x*6364136223846793005 + 1442695040888963407
		keys[i] = x
	}
	return keys
}

// sink 防编译器消除结果
func sink(b *testing.B, acc uint64) {
	if acc == 0 {
		b.Fatal("bench sink: unreachable")
	}
}

// ---------------- 字符串 key 单查询 ----------------

func BenchmarkGetNode_StringKey(b *testing.B) {
	shortKeys := benchStringKeys(benchStringPool)
	longKeys := benchLongKeys(benchStringPool)

	for _, nc := range benchNodeCounts {
		h := benchBuildHash(b, nc, benchVirtuals)
		// 短 key（约 13 字节）
		b.Run(fmt.Sprintf("short/nodes=%d", nc), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(shortKeys[0])))
			b.ResetTimer()
			var acc uint64
			for i := 0; i < b.N; i++ {
				acc += uint64(h.GetNode(shortKeys[i%len(shortKeys)]).GetId())
			}
			sink(b, acc)
		})
		// 长 key（约 128 字节，放大哈希计算）
		b.Run(fmt.Sprintf("long/nodes=%d", nc), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(longKeys[0])))
			b.ResetTimer()
			var acc uint64
			for i := 0; i < b.N; i++ {
				acc += uint64(h.GetNode(longKeys[i%len(longKeys)]).GetId())
			}
			sink(b, acc)
		})
	}
}

// ---------------- uint64 key 单查询 ----------------

func BenchmarkGetNodeByHash(b *testing.B) {
	keys := benchUintKeys(benchUintPool)

	for _, nc := range benchNodeCounts {
		h := benchBuildHash(b, nc, benchVirtuals)

		// 顺序 key（贴近真实递增 ID）
		b.Run(fmt.Sprintf("sequential/nodes=%d", nc), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(8)
			b.ResetTimer()
			var acc uint64
			for i := 0; i < b.N; i++ {
				acc += uint64(h.GetNodeByHash(uint64(i)).GetId())
			}
			sink(b, acc)
		})

		// 伪随机 key
		b.Run(fmt.Sprintf("random/nodes=%d", nc), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(8)
			b.ResetTimer()
			var acc uint64
			for i := 0; i < b.N; i++ {
				acc += uint64(h.GetNodeByHash(keys[i%len(keys)]).GetId())
			}
			sink(b, acc)
		})
	}
}

// ---------------- 主备链查询（故障转移） ----------------

func BenchmarkGetNodesByKey(b *testing.B) {
	keys := benchStringKeys(benchStringPool)
	chainLens := []int{2, 3, 5} // 主 + 1/2/4 备份

	for _, nc := range benchNodeCounts {
		h := benchBuildHash(b, nc, benchVirtuals)
		for _, cl := range chainLens {
			b.Run(fmt.Sprintf("chain=%d/nodes=%d", cl, nc), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				var acc uint64
				for i := 0; i < b.N; i++ {
					chain := h.GetNodesByKey(keys[i%len(keys)], cl)
					acc += uint64(len(chain))
				}
				sink(b, acc)
			})
		}
	}
}

// ---------------- 并发读（锁竞争压力，配合 -cpu=1,8,16,32 观察扩展性） ----------------

func BenchmarkParallelGetNode_StringKey(b *testing.B) {
	h := benchBuildHash(b, 1000, benchVirtuals) // 大环 1000 节点 = 15 万槽位
	keys := benchStringKeys(benchStringPool)
	benchSinkTotal.Store(0)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var acc uint64
		i := 0
		for pb.Next() {
			acc += uint64(h.GetNode(keys[i%len(keys)]).GetId())
			i++
		}
		benchSinkTotal.Add(acc)
	})

	if benchSinkTotal.Load() == 0 {
		b.Fatal("bench sink: zero result")
	}
}

func BenchmarkParallelGetNodeByHash(b *testing.B) {
	h := benchBuildHash(b, 1000, benchVirtuals)
	benchSinkTotal.Store(0)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var acc uint64
		i := uint64(0)
		for pb.Next() {
			acc += uint64(h.GetNodeByHash(i).GetId())
			i++
		}
		benchSinkTotal.Add(acc)
	})

	if benchSinkTotal.Load() == 0 {
		b.Fatal("bench sink: zero result")
	}
}

func BenchmarkParallelGetNodesByKey(b *testing.B) {
	h := benchBuildHash(b, 1000, benchVirtuals)
	keys := benchStringKeys(benchStringPool)
	benchSinkTotal.Store(0)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		var acc uint64
		i := 0
		for pb.Next() {
			acc += uint64(len(h.GetNodesByKey(keys[i%len(keys)], 3)))
			i++
		}
		benchSinkTotal.Add(acc)
	})

	if benchSinkTotal.Load() == 0 {
		b.Fatal("bench sink: zero result")
	}
}

// ---------------- 写路径（节点上下线成本，低频操作） ----------------

// hotSwap 在固定规模环上做一次“移除已有节点 + 同 id 重新加入”（模拟热插拔），环规模恒定
func hotSwap(b *testing.B, h *Hash[uint32, *testNode], id uint32) {
	b.Helper()
	if h.RemoveNode(id) == nil {
		b.Fatalf("RemoveNode(%d) returned nil", id)
	}
	if err := h.AddNode(id, &testNode{id: id, name: "hot"}); err != nil {
		b.Fatalf("AddNode(%d): %v", id, err)
	}
}

func BenchmarkAddNode(b *testing.B) {
	// 测量 虚拟点生成 + 冲突预检 + 线性归并（新节点加入满环）
	for _, nc := range []int{100, 1000, 2000} {
		b.Run(fmt.Sprintf("ring=%d", nc*benchVirtuals), func(b *testing.B) {
			base := benchBuildHash(b, nc, benchVirtuals)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hotSwap(b, base, uint32(1+i%nc)) // 热插拔同 id，环规模稳定
			}
		})
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	// 测量整环线性过滤删除某一节点的槽位
	for _, nc := range []int{100, 1000, 2000} {
		b.Run(fmt.Sprintf("ring=%d", nc*benchVirtuals), func(b *testing.B) {
			base := benchBuildHash(b, nc, benchVirtuals)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hotSwap(b, base, uint32(1+i%nc))
			}
		})
	}
}
