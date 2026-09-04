package consistent

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

// testNode 测试用最小节点实现
type testNode struct {
	id   uint32
	name string
}

func (n *testNode) GetId() uint32 { return n.id }

// newTestNodesFrom 创建 id 从 start 开始、共 count 个节点
func newTestNodesFrom(start, count int) []*testNode {
	nodes := make([]*testNode, 0, count)
	for i := start; i < start+count; i++ {
		nodes = append(nodes, &testNode{id: uint32(i), name: fmt.Sprintf("node-%d", i)})
	}
	return nodes
}

const testVirtuals = 100

// addAll 批量添加节点
func addAll(t *testing.T, h *Hash[uint32, *testNode], nodes []*testNode) {
	t.Helper()
	for _, n := range nodes {
		if err := h.AddNode(n.id, n); err != nil {
			t.Fatalf("AddNode(%d): %v", n.id, err)
		}
	}
}

// ---------- 分布均衡性 ----------

func TestDistributionBalanceUint64Key(t *testing.T) {
	const (
		nodeCount = 100
		keyCount  = 200000
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, nodeCount))

	counts := make(map[uint32]int, nodeCount)
	for i := uint64(0); i < keyCount; i++ {
		n := h.GetNodeByHash(i)
		if n == nil {
			t.Fatal("非空环 GetNodeByHash 返回 nil")
		}
		counts[n.GetId()]++
	}

	if len(counts) != nodeCount {
		t.Fatalf("期望 %d 个节点都分到 key，实际只有 %d 个", nodeCount, len(counts))
	}

	mean := float64(keyCount) / float64(nodeCount)
	var minV, maxV int
	first := true
	for _, c := range counts {
		if first {
			minV, maxV, first = c, c, false
			continue
		}
		if c < minV {
			minV = c
		}
		if c > maxV {
			maxV = c
		}
	}
	t.Logf("counts=%v mean=%.0f min=%d max=%d", counts, mean, minV, maxV)

	// 一致性哈希均衡性理论波动约 ±1/√virtuals≈10%（100虚拟节点），最差节点可到 ~25%。
	// ±30% 阈值稳健不误报，但仍能捕获真实缺陷（虚拟点成簇/节点饿死时偏差可达 3 倍以上）
	if float64(minV) < mean*0.7 {
		t.Errorf("均衡性差：min=%d < mean*0.7=%.0f", minV, mean*0.7)
	}
	if float64(maxV) > mean*1.3 {
		t.Errorf("均衡性差：max=%d > mean*1.3=%.0f", maxV, mean*1.3)
	}
}

func TestDistributionBalanceStringKey(t *testing.T) {
	const (
		nodeCount = 120
		keyCount  = 200000
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, nodeCount))

	counts := make(map[uint32]int, nodeCount)
	for i := 0; i < keyCount; i++ {
		n := h.GetNode(fmt.Sprintf("player-%d", i))
		if n == nil {
			t.Fatal("非空环 GetNode 返回 nil")
		}
		counts[n.GetId()]++
	}

	if len(counts) != nodeCount {
		t.Fatalf("期望 %d 个节点都分到 key，实际 %d", nodeCount, len(counts))
	}

	mean := float64(keyCount) / float64(nodeCount)
	var minV, maxV int
	first := true
	for _, c := range counts {
		if first {
			minV, maxV, first = c, c, false
			continue
		}
		if c < minV {
			minV = c
		}
		if c > maxV {
			maxV = c
		}
	}
	t.Logf("mean=%.0f min=%d max=%d", mean, minV, maxV)

	// 理论波动约 ±1/√virtuals，最差节点 ~25%；±30% 稳健且能捕获成簇类缺陷
	if float64(minV) < mean*0.7 || float64(maxV) > mean*1.3 {
		t.Errorf("均衡性差：min=%d max=%d (mean=%.0f)", minV, maxV, mean)
	}
}

// ---------- 增删节点后最小重映射 ----------

func TestMinimalRemappingOnRemove(t *testing.T) {
	const (
		nodeCount = 10
		keyCount  = 100000
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	nodes := newTestNodesFrom(1, nodeCount)
	addAll(t, h, nodes)

	ownerBefore := make(map[uint64]uint32, keyCount)
	for i := uint64(0); i < keyCount; i++ {
		ownerBefore[i] = h.GetNodeByHash(i).GetId()
	}

	removed := nodes[nodeCount-1]
	h.RemoveNode(removed.id)

	remapped := 0
	for i := uint64(0); i < keyCount; i++ {
		n := h.GetNodeByHash(i)
		if n == nil {
			t.Fatal("移除后路由到 nil")
		}
		if n.GetId() == removed.id {
			t.Fatalf("key=%d 仍路由到已移除节点 %d", i, removed.id)
		}
		if n.GetId() != ownerBefore[i] {
			remapped++
		}
	}

	// 理论上只有原属被移除节点的 ~1/N key 迁移；若接近全量重排即 bug
	expected := float64(keyCount) / float64(nodeCount)
	upper := int(expected*1.5) + 1
	t.Logf("移除1节点后重映射=%d，期望≈%.0f", remapped, expected)
	if remapped > upper {
		t.Errorf("移除 1 个节点后重映射过多：%d > %d", remapped, upper)
	}
}

func TestMinimalRemappingOnAdd(t *testing.T) {
	const (
		initialNodes = 9
		keyCount     = 100000
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, initialNodes))

	ownerBefore := make(map[uint64]uint32, keyCount)
	for i := uint64(0); i < keyCount; i++ {
		ownerBefore[i] = h.GetNodeByHash(i).GetId()
	}

	// 新增第 10 个节点
	addAll(t, h, newTestNodesFrom(initialNodes+1, 1))

	remapped := 0
	for i := uint64(0); i < keyCount; i++ {
		n := h.GetNodeByHash(i)
		if n == nil {
			t.Fatal("路由到 nil")
		}
		if n.GetId() != ownerBefore[i] {
			remapped++
		}
	}

	// 新增 1 个节点，只有约 1/(N+1) 的 key 迁移到新节点
	expected := float64(keyCount) / float64(initialNodes+1)
	upper := int(expected*1.5) + 1
	t.Logf("新增1节点后重映射=%d，期望≈%.0f", remapped, expected)
	if remapped > upper {
		t.Errorf("新增 1 个节点后重映射过多：%d > %d", remapped, upper)
	}
}

// ---------- 主备链一致性 ----------

func TestBackupChainConsistency(t *testing.T) {
	const (
		nodeCount = 8
		chainLen  = 3
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, nodeCount))

	for i := 0; i < 20000; i++ {
		key := fmt.Sprintf("key-%d", i)

		chain1 := h.GetNodesByKey(key, chainLen)
		if len(chain1) != chainLen {
			t.Fatalf("key=%s chain 长度=%d 期望 %d", key, len(chain1), chainLen)
		}

		// 确定性：多次查询结果一致
		chain2 := h.GetNodesByKey(key, chainLen)
		for j := 0; j < chainLen; j++ {
			if chain1[j].GetId() != chain2[j].GetId() {
				t.Fatalf("key=%s 两次查询结果不一致", key)
			}
		}

		// 链首 == 主节点
		if got := h.GetNode(key); got == nil || chain1[0].GetId() != got.GetId() {
			t.Errorf("key=%s 主节点不一致：chain[0]=%d", key, chain1[0].GetId())
		}

		// 链内节点去重且全部在线
		seen := make(map[uint32]bool, chainLen)
		for _, n := range chain1 {
			if seen[n.GetId()] {
				t.Errorf("key=%s chain 内出现重复节点 %d", key, n.GetId())
			}
			seen[n.GetId()] = true
			if h.GetNodeByKey(n.GetId()) == nil {
				t.Errorf("key=%s chain 含已下线节点 %d", key, n.GetId())
			}
		}
	}
}

func TestBackupPromotionAfterRemove(t *testing.T) {
	const (
		nodeCount = 8
		chainLen  = 3
	)

	h := NewHash[uint32, *testNode](testVirtuals)
	nodes := newTestNodesFrom(1, nodeCount)
	addAll(t, h, nodes)
	removed := nodes[0] // id=1

	// 记录主节点为被移除节点的 key 及其旧主备链
	before := make(map[string][]uint32)
	for i := 0; i < 20000; i++ {
		key := fmt.Sprintf("key-%d", i)
		chain := h.GetNodesByKey(key, chainLen)
		if len(chain) != chainLen {
			t.Fatalf("chain len=%d", len(chain))
		}
		if chain[0].GetId() == removed.id {
			before[key] = []uint32{chain[0].GetId(), chain[1].GetId(), chain[2].GetId()}
		}
	}
	if len(before) == 0 {
		t.Fatal("没有 key 以被移除节点为主节点，测试无效")
	}
	t.Logf("受影响 key=%d", len(before))

	h.RemoveNode(removed.id)

	for key, old := range before {
		chain := h.GetNodesByKey(key, chainLen)
		if len(chain) != chainLen {
			t.Fatalf("key=%s 移除后 chain len=%d", key, len(chain))
		}
		if chain[0].GetId() == removed.id {
			t.Fatalf("key=%s 仍路由到已移除节点", key)
		}
		// 旧主被移除后：旧第一备份应提升为主，旧第二备份顺延为第二
		if chain[0].GetId() != old[1] {
			t.Errorf("key=%s 期望旧备份 %d 提升为主，实际主=%d", key, old[1], chain[0].GetId())
		}
		if chain[1].GetId() != old[2] {
			t.Errorf("key=%s 期望旧第二备份 %d 顺延，实际第二=%d", key, old[2], chain[1].GetId())
		}
	}
}

// ---------- 基础 API 与边界 ----------

func TestAddDuplicateRejected(t *testing.T) {
	h := NewHash[uint32, *testNode](testVirtuals)
	if err := h.AddNode(7, &testNode{id: 7, name: "a"}); err != nil {
		t.Fatalf("首次 AddNode 失败: %v", err)
	}
	if err := h.AddNode(7, &testNode{id: 7, name: "b"}); err == nil {
		t.Fatal("重复 id 添加应返回错误")
	}
	if h.GetNodeCount() != 1 || h.GetVirtualNodeCount() != testVirtuals {
		t.Fatalf("重复添加污染状态: nodes=%d virtuals=%d", h.GetNodeCount(), h.GetVirtualNodeCount())
	}
}

func TestRemoveNonExistent(t *testing.T) {
	h := NewHash[uint32, *testNode](testVirtuals)
	if got := h.RemoveNode(123); got != nil {
		t.Fatalf("移除不存在节点应返回 nil，实际 %v", got)
	}
}

func TestEmptyRingReturnsNil(t *testing.T) {
	h := NewHash[uint32, *testNode](testVirtuals)
	if h.GetNode("x") != nil {
		t.Fatal("空环 GetNode 应返回 nil")
	}
	if h.GetNodeByHash(1) != nil {
		t.Fatal("空环 GetNodeByHash 应返回 nil")
	}
	if h.GetNodesByKey("x", 2) != nil {
		t.Fatal("空环 GetNodesByKey 应返回 nil")
	}
}

func TestVirtualNodeCountConsistency(t *testing.T) {
	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, 4))

	if h.GetNodeCount() != 4 || h.GetVirtualNodeCount() != 4*testVirtuals {
		t.Fatalf("节点数=%d 虚拟节点数=%d，期望 4 / %d（出现重复说明环被污染）",
			h.GetNodeCount(), h.GetVirtualNodeCount(), 4*testVirtuals)
	}

	h.RemoveNode(2)
	h.RemoveNode(4)
	if h.GetNodeCount() != 2 || h.GetVirtualNodeCount() != 2*testVirtuals {
		t.Fatalf("删除后节点数=%d 虚拟节点数=%d，期望 2 / %d",
			h.GetNodeCount(), h.GetVirtualNodeCount(), 2*testVirtuals)
	}

	// 删除后可重新加入，环恢复一致
	if err := h.AddNode(2, &testNode{id: 2, name: "n2-again"}); err != nil {
		t.Fatalf("重新添加失败: %v", err)
	}
	if h.GetNodeCount() != 3 || h.GetVirtualNodeCount() != 3*testVirtuals {
		t.Fatalf("重加后节点数=%d 虚拟节点数=%d，期望 3 / %d",
			h.GetNodeCount(), h.GetVirtualNodeCount(), 3*testVirtuals)
	}
}

func TestUpdateNodeKeepsRouting(t *testing.T) {
	h := NewHash[uint32, *testNode](testVirtuals)
	addAll(t, h, newTestNodesFrom(1, 3))

	// 记录映射
	owner := make(map[uint64]uint32, 5000)
	for i := uint64(0); i < 5000; i++ {
		owner[i] = h.GetNodeByHash(i).GetId()
	}

	// 同 id 更新元数据（如 ip/端口变更），不应改变路由
	if err := h.UpdateNode(1, &testNode{id: 1, name: "node-1-updated"}); err != nil {
		t.Fatalf("UpdateNode: %v", err)
	}
	if got := h.GetNodeByKey(1); got == nil || got.name != "node-1-updated" {
		t.Fatalf("UpdateNode 后 GetNodeByKey 未返回新节点: %+v", got)
	}
	for i := uint64(0); i < 5000; i++ {
		if n := h.GetNodeByHash(i); n == nil || n.GetId() != owner[i] {
			t.Fatalf("UpdateNode 导致路由变化 key=%d", i)
		}
	}

	// 更新不存在的节点应报错
	if err := h.UpdateNode(999, &testNode{id: 999, name: "x"}); err == nil {
		t.Fatal("更新不存在节点应报错")
	}
}

// ---------- 并发安全（配合 go test -race） ----------

func TestConcurrentReadDuringMutation(t *testing.T) {
	h := NewHash[uint32, *testNode](50)
	addAll(t, h, newTestNodesFrom(1, 5)) // 常驻节点，保证环非空
	extra := newTestNodesFrom(6, 10)     // 反复上下线的节点

	done := make(chan struct{})
	var wg sync.WaitGroup

	// 写方：反复上下线节点
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)
		for iter := 0; iter < 5000; iter++ {
			n := extra[iter%len(extra)]
			if h.GetNodeByKey(n.id) == nil {
				_ = h.AddNode(n.id, n)
			} else {
				h.RemoveNode(n.id)
			}
		}
	}()

	// 读方：持续路由
	const readers = 8
	for g := 0; g < readers; g++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				h.GetNodeByHash(seed)
				h.GetNodesByKey(fmt.Sprintf("k-%d", seed), 2)
				seed = seed*6364136223846793005 + 1442695040888963407
			}
		}(uint64(g+1) * 2654435761)
	}

	wg.Wait()

	// 结束后常驻节点仍在，路由可用
	if h.GetNodeCount() < 5 {
		t.Fatalf("并发结束后常驻节点丢失: %d", h.GetNodeCount())
	}
	if h.GetNodeByHash(42) == nil {
		t.Fatal("并发结束后路由返回 nil")
	}
}

// ---------- 大规模无冲突哨兵（B 的 canary） ----------

// 2000 节点 × 100 虚拟点 = 20 万个环位置。AddNode 内置 B 冲突预检：任何虚拟点与
// 已有节点冲突都会让 AddNode 返回错误（addAll 已 Fatalf），因此能走到断言处即证明
// 20 万虚拟点零冲突；环槽位数 == 节点数×virtuals 进一步锁定无重复/丢失。
func TestNoCollisionAtScale(t *testing.T) {
	const (
		nodeCount = 2000
		virtuals  = 100
	)

	h := NewHash[uint32, *testNode](virtuals)
	addAll(t, h, newTestNodesFrom(1, nodeCount))

	if h.GetNodeCount() != nodeCount {
		t.Fatalf("节点数=%d 期望 %d", h.GetNodeCount(), nodeCount)
	}
	if got := h.GetVirtualNodeCount(); got != nodeCount*virtuals {
		t.Fatalf("虚拟槽位数=%d 期望 %d：存在冲突/重复/丢失", got, nodeCount*virtuals)
	}

	// 环必须有序
	for i := 1; i < len(h.hashRing); i++ {
		if h.hashRing[i-1].hash > h.hashRing[i].hash {
			t.Fatalf("环未有序: [%d]=%#x > [%d]=%#x", i-1, h.hashRing[i-1].hash, i, h.hashRing[i].hash)
		}
	}

	// 每个节点至少分到 key（无饥饿；N=2000、K=200000 时 P(零)≈e^-100）
	counts := make(map[uint32]int, nodeCount)
	for i := uint64(0); i < 200000; i++ {
		n := h.GetNodeByHash(i)
		if n == nil {
			t.Fatal("非空环路由返回 nil")
		}
		counts[n.GetId()]++
	}
	if len(counts) != nodeCount {
		t.Fatalf("期望 %d 个节点都分到 key，实际 %d", nodeCount, len(counts))
	}
}

// ---------- 扩容最小迁移量（运行期 AddNode，标准做法验证） ----------

// 与 TestStaticExpansionMinimalRemapping 对偶：锁版 Hash 支持运行期扩容——
// 直接对运行中的环 AddNode(d)。验证一致性哈希扩容的标准做法：
//   - 迁移只流向新节点 d（跨旧节点零搬移，旧节点之间不互相搬家）；
//   - 总迁移 = 1/(N+1)（理论最小）；
//   - 每个旧节点被 d 均摊抢走约 1/(N(N+1)) 的地盘。
func TestExpansionMinimalRemapping(t *testing.T) {
	const (
		virtuals = 100
		keyCount = 1000000
		newID    = 9999 // 新增物理节点 d 的 id（与旧节点 id 区间错开）
	)

	for _, tt := range []struct {
		name     string
		oldNodes int
	}{
		{"3->4", 3}, // 期望总迁移 1/4=25%，每旧节点 1/12≈8.33%
		{"8->9", 8}, // 期望总迁移 1/9≈11.1%，每旧节点 1/72≈1.39%
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHash[uint32, *testNode](virtuals)
			addAll(t, h, newTestNodesFrom(1, tt.oldNodes))

			expectedTotal := 1.0 / float64(tt.oldNodes+1) // 理论最小总迁移 1/(N+1)
			expectedPerOld := expectedTotal / float64(tt.oldNodes)

			// 记录扩容前每个 key 的归属
			ownerBefore := make([]uint32, keyCount)
			for i := uint64(0); i < keyCount; i++ {
				ownerBefore[i] = h.GetNodeByHash(i).GetId()
			}

			// 运行期新增节点 d
			if err := h.AddNode(newID, &testNode{id: newID, name: "d"}); err != nil {
				t.Fatalf("AddNode(d): %v", err)
			}

			migrated := 0
			crossOld := 0 // 旧节点之间互相搬移的 key 数（应为 0）
			lostBy := make(map[uint32]int, tt.oldNodes)
			for i := uint64(0); i < keyCount; i++ {
				before := ownerBefore[i]
				after := h.GetNodeByHash(i).GetId()
				if before == after {
					continue
				}
				migrated++
				if after == newID {
					lostBy[before]++ // 旧节点 before 被 d 抢走的 key
				} else {
					crossOld++ // 迁移未流向新节点 → 旧节点互抢，违反标准做法
				}
			}

			rate := float64(migrated) / float64(keyCount)

			// 强特征 1：迁移只流向新节点，旧节点之间零搬移
			if crossOld != 0 {
				t.Errorf("扩容引起旧节点间迁移 %d 个 key（迁移必须只流向新节点）", crossOld)
			}
			// 强特征 2：总迁移 = 1/(N+1)，理论最小
			if rate < expectedTotal*0.8 || rate > expectedTotal*1.2 {
				t.Errorf("总迁移=%.4f%% 偏离理论最小 %.2f%% 过大", rate*100, expectedTotal*100)
			}

			// 强特征 3：每个旧节点被 d 均摊抢走约 1/(N(N+1))，无节点被过度劫掠
			rates := make([]string, 0, tt.oldNodes)
			for id := uint32(1); id <= uint32(tt.oldNodes); id++ {
				per := float64(lostBy[id]) / float64(keyCount)
				rates = append(rates, fmt.Sprintf("%d:%.3f%%", id, per*100))
				if per < expectedPerOld*0.4 || per > expectedPerOld*2.2 {
					t.Errorf("旧节点 %d 丢失 %.4f%% (期望 %.2f%%)，偏离均摊", id, per*100, expectedPerOld*100)
				}
			}
			t.Logf("总迁移=%d (%.4f%%, 理论 %.2f%%) 跨旧节点=%d lostBy={%s}",
				migrated, rate*100, expectedTotal*100, crossOld, strings.Join(rates, " "))
		})
	}
}
