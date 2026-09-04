package consistent

import (
	"fmt"
	"sort"
	"sync/atomic"
)

// StaticHash 一次性构建、运行期只读的「完全无锁」一致性哈希。
// A：节点 id 类型（comparable）；T：节点负载类型（任意）。
//
// 与 Hash[A, T] 的关系：播种方式与环结构完全一致（同一节点 id 集合得到
// 完全相同的环布局与路由结果），可在一处切换实现而无数据归属变化；
// 区别仅在同步策略：
//   - Hash      ：读路径每次加 RWMutex，支持运行期动态增删节点（成本换灵活性）；
//   - StaticHash：读路径完全无锁、无原子、无堆分配（零同步开销），
//     代价是构建完成（Freeze）后不允许任何写操作。
//
// 适用场景：服务启动时一次性加载全部节点，运行期不会新增/移除/变更节点；
// 任何节点变更（新增、上下线、扩缩容）都必须重启服务后重新构建。
//
// 用法：
//
//	h := consistent.NewStaticHash[uint32, *Node](150)
//	h.AddNode(1, n1)
//	h.AddNode(2, n2)
//	h.Freeze() // 构建完成，进入只读阶段（对外提供服务前调用）
//
//	// 此后 h 可被任意多个 goroutine 并发读取，全程无锁
//	owner := h.GetNodeByHash(playerID)
//	chain := h.GetNodesByKey("player-10001", 3)
//
// 并发安全模型：
//   - 构建阶段（AddNode）必须在启动流程中完成，写入发生在对外发布之前
//     （同一 goroutine 顺序执行，或经通道/启动同步建立 happens-before）；
//   - Freeze 之后所有字段不再写入。Go 语言保证多 goroutine 并发只读
//     slice 与 map 是安全的，因此读路径不需要任何同步原语；
//   - 误在 Freeze 后调用 AddNode 会 panic（frozen 为原子检查，误用也呈确定行为）。
type StaticHash[A comparable, T any] struct {
	hashRing []ringSlot[A, T] // 哈希环（构建期写入，Freeze 后只读）
	nodes    map[A]T          // 真实节点索引（同上）
	virtuals int              // 每个真实节点的虚拟节点数
	frozen   atomic.Bool      // Freeze 标记：仅写路径检查，读路径不触碰
}

// NewStaticHash 创建 StaticHash（构建期对象，AddNode 后需 Freeze）
func NewStaticHash[A comparable, T any](virtuals int) *StaticHash[A, T] {
	if virtuals <= 0 {
		virtuals = 150 // 默认虚拟节点数
	}
	return &StaticHash[A, T]{
		hashRing: make([]ringSlot[A, T], 0),
		nodes:    make(map[A]T),
		virtuals: virtuals,
	}
}

// Freeze 冻结哈希：构建完成后调用，此后只读。重复调用安全。
// 返回接收者自身，便于链式。Freeze 后调用 AddNode 将 panic。
func (h *StaticHash[A, T]) Freeze() *StaticHash[A, T] {
	h.frozen.Store(true)
	return h
}

// AddNode 添加节点（仅构建期、Freeze 前调用，无锁——构建假定单 goroutine）。
// nodeId 播种与 Hash.AddNode 完全一致：第 i 个虚拟点环位置 =
// CalcHashVirtualKey(fmt.Sprint(nodeId), i)。节点 id 重复会报错。
// Freeze 后调用会 panic。
func (h *StaticHash[A, T]) AddNode(nodeId A, node T) error {
	if h.frozen.Load() {
		panic("consistent: StaticHash.AddNode called after Freeze (hash is read-only)")
	}

	// 检查节点是否已存在
	if _, exists := h.nodes[nodeId]; exists {
		return fmt.Errorf("node already exists: %v (id=%v)", node, nodeId)
	}

	// 先离线算好本节点全部虚拟点并做冲突预检——与已有节点冲突立即报错，
	// 此时尚未写入任何状态（无半截数据）；环槽位自带节点 id，即便漏检也不会损坏结构。
	seed := fmt.Sprint(nodeId) // 播种串：整型 id 时为十进制字符串
	slots := make([]ringSlot[A, T], 0, h.virtuals)
	seen := make(map[uint64]struct{}, h.virtuals) // 同节点自身去重
	for i := 0; i < h.virtuals; i++ {
		vk := CalcHashVirtualKey(seed, i)
		if _, dup := seen[vk]; dup {
			continue // 同节点内重复（需 64 位真碰撞才可能）：跳过，避免环冗余
		}
		seen[vk] = struct{}{}
		if hasRingPoint(h.hashRing, vk) {
			return fmt.Errorf("virtual hash collision: node %v conflicts with existing ring point %#x", nodeId, vk)
		}
		slots = append(slots, ringSlot[A, T]{hash: vk, a: nodeId, node: node})
	}

	// 添加真实节点
	h.nodes[nodeId] = node

	// 新节点槽位排序后与现有环线性归并，避免每次全量重排
	sort.Slice(slots, func(i, j int) bool { return slots[i].hash < slots[j].hash })
	if len(h.hashRing) == 0 {
		h.hashRing = slots
		return nil
	}

	merged := make([]ringSlot[A, T], 0, len(h.hashRing)+len(slots))
	i, j := 0, 0
	for i < len(h.hashRing) && j < len(slots) {
		if h.hashRing[i].hash <= slots[j].hash {
			merged = append(merged, h.hashRing[i])
			i++
		} else {
			merged = append(merged, slots[j])
			j++
		}
	}
	merged = append(merged, h.hashRing[i:]...)
	merged = append(merged, slots[j:]...)
	h.hashRing = merged
	return nil
}

// searchFirstGE 在有序环上二分查找第一个 hash >= target 的槽位下标
// （返回值范围 [0, len(ring)]，等于 len 表示无满足项、需回绕到环首）。
// 手写二分替代 sort.Search：无闭包、无间接调用，读路径零额外开销。
func searchFirstGE[A comparable, T any](ring []ringSlot[A, T], target uint64) int {
	lo, hi := 0, len(ring)
	for lo < hi {
		mid := int(uint(lo+hi) >> 1) // 防溢出写法
		if ring[mid].hash < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// hasRingPoint 判断环上是否已存在相同 hash 的槽位（二分查找）
func hasRingPoint[A comparable, T any](ring []ringSlot[A, T], hash uint64) bool {
	idx := searchFirstGE(ring, hash)
	return idx < len(ring) && ring[idx].hash == hash
}

// GetNode 根据 string key 获取对应的节点（无锁，Freeze 后并发安全）
func (h *StaticHash[A, T]) GetNode(key string) T {
	if len(h.hashRing) == 0 {
		var zero T
		return zero
	}
	hash := hashKey(key)
	idx := searchFirstGE(h.hashRing, hash)
	if idx == len(h.hashRing) {
		idx = 0 // 环形回绕
	}
	return h.hashRing[idx].node
}

// GetNodeByHash 根据 uint64 key 获取对应的节点（无锁，Freeze 后并发安全）
func (h *StaticHash[A, T]) GetNodeByHash(key uint64) T {
	if len(h.hashRing) == 0 {
		var zero T
		return zero
	}
	hash := hashUint64(key)
	idx := searchFirstGE(h.hashRing, hash)
	if idx == len(h.hashRing) {
		idx = 0
	}
	return h.hashRing[idx].node
}

// GetNodeByKey 根据节点 id 获取节点（无锁；map 只读并发安全）
func (h *StaticHash[A, T]) GetNodeByKey(nodeID A) T {
	return h.nodes[nodeID]
}

// GetNodes 获取所有节点（无锁；返回副本）
func (h *StaticHash[A, T]) GetNodes() []T {
	nodes := make([]T, 0, len(h.nodes))
	for _, node := range h.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodeCount 获取节点数量（无锁）
func (h *StaticHash[A, T]) GetNodeCount() int {
	return len(h.nodes)
}

// GetVirtualNodeCount 获取虚拟节点数量（即哈希环槽位数；无锁）
func (h *StaticHash[A, T]) GetVirtualNodeCount() int {
	return len(h.hashRing)
}

// GetNodesByKey 获取某个 key 的所有备份节点（用于故障转移；无锁）
func (h *StaticHash[A, T]) GetNodesByKey(key string, count int) []T {
	if len(h.hashRing) == 0 || count <= 0 {
		return nil
	}

	hash := hashKey(key)
	// 起点 = 首个 hash >= key 的槽位；越界即回绕（环形）
	base := searchFirstGE(h.hashRing, hash) % len(h.hashRing)

	nodes := make([]T, 0, count)
	seen := make(map[A]bool)

	// 沿环顺时针取前 count 个不同真实节点（按槽位所属节点 id 判重）
	for i := 0; i < len(h.hashRing) && len(nodes) < count; i++ {
		s := h.hashRing[(base+i)%len(h.hashRing)]
		if !seen[s.a] {
			seen[s.a] = true
			nodes = append(nodes, s.node)
		}
	}
	return nodes
}
