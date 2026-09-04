package consistent

import (
	"fmt"
	"sort"
	"sync"
)

// ringSlot 哈希环槽位：自带所属节点 id（a）与节点值。即使两个虚拟点的哈希碰撞，
// 两个槽位也能各自保留、互不覆盖；删除时也只移除属于该 id 的槽位，结构上对冲突免疫。
type ringSlot[A comparable, T any] struct {
	hash uint64
	a    A // 槽位所属真实节点的 id
	node T // 槽位所属节点
}

// Hash 一致性哈希实现
// A：节点 id 类型（comparable，用作 nodes 索引与环槽位的归属标识）
// T：节点负载类型（任意）
type Hash[A comparable, T any] struct {
	mu       sync.RWMutex
	nodes    map[A]T          // 真实节点（按节点 id 索引）
	hashRing []ringSlot[A, T] // 哈希环（按 hash 升序，槽位自带节点 id）
	virtuals int              // 每个真实节点的虚拟节点数
}

// NewHash 创建一致性哈希实例
func NewHash[A comparable, T any](virtuals int) *Hash[A, T] {
	if virtuals <= 0 {
		virtuals = 150 // 默认虚拟节点数
	}
	return &Hash[A, T]{
		nodes:    make(map[A]T),
		hashRing: make([]ringSlot[A, T], 0),
		virtuals: virtuals,
	}
}

// AddNode 添加节点。nodeId 为该节点唯一 id（环槽位以它标识归属并去重），
// 该节点的第 i 个虚拟点环位置 = hash(fmt.Sprint(nodeId) + ":" + i)。
// 整型 id 的播种布局与 CalcHashVirtualNode 一致（十进制无前导零）。
func (ch *Hash[A, T]) AddNode(nodeId A, node T) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否已存在
	if _, exists := ch.nodes[nodeId]; exists {
		return fmt.Errorf("node already exists: %v (id=%v)", node, nodeId)
	}

	// B：先离线算好本节点全部虚拟点并做冲突预检——若与已有节点冲突立即报错，
	// 此时尚未写入任何状态（无半截数据）；环槽位自带节点 id（C），即便漏检也不会损坏结构。
	seed := fmt.Sprint(nodeId) // 播种串：整型 id 时为十进制字符串
	slots := make([]ringSlot[A, T], 0, ch.virtuals)
	seen := make(map[uint64]struct{}, ch.virtuals) // 同节点自身去重
	for i := 0; i < ch.virtuals; i++ {
		vk := CalcHashVirtualKey(seed, i)
		if _, dup := seen[vk]; dup {
			continue // 同节点内重复（需 64 位真碰撞才可能）：跳过，避免环冗余
		}
		seen[vk] = struct{}{}
		if ch.hasRingPoint(vk) {
			return fmt.Errorf("virtual hash collision: node %v conflicts with existing ring point %#x", nodeId, vk)
		}
		slots = append(slots, ringSlot[A, T]{hash: vk, a: nodeId, node: node})
	}

	// 添加真实节点
	ch.nodes[nodeId] = node

	// C：新节点槽位排序后与现有环线性归并，避免每次全量重排
	sort.Slice(slots, func(i, j int) bool { return slots[i].hash < slots[j].hash })
	if len(ch.hashRing) == 0 {
		ch.hashRing = slots
		return nil
	}

	merged := make([]ringSlot[A, T], 0, len(ch.hashRing)+len(slots))
	i, j := 0, 0
	for i < len(ch.hashRing) && j < len(slots) {
		if ch.hashRing[i].hash <= slots[j].hash {
			merged = append(merged, ch.hashRing[i])
			i++
		} else {
			merged = append(merged, slots[j])
			j++
		}
	}
	merged = append(merged, ch.hashRing[i:]...)
	merged = append(merged, slots[j:]...)
	ch.hashRing = merged
	return nil
}

// hasRingPoint 判断环上是否已存在相同 hash 的槽位（二分查找）
func (ch *Hash[A, T]) hasRingPoint(hash uint64) bool {
	idx := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i].hash >= hash
	})
	return idx < len(ch.hashRing) && ch.hashRing[idx].hash == hash
}

// RemoveNode 移除节点（按节点 id）
func (ch *Hash[A, T]) RemoveNode(nodeID A) T {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否存在
	node, exists := ch.nodes[nodeID]
	if !exists {
		var zero T
		return zero
	}

	// 移除真实节点
	delete(ch.nodes, nodeID)

	// C：仅移除属于该节点的槽位（槽位自带节点 id，绝不误删他人、不留孤儿）
	newRing := make([]ringSlot[A, T], 0, len(ch.hashRing))
	for _, s := range ch.hashRing {
		if s.a != nodeID {
			newRing = append(newRing, s)
		}
	}
	ch.hashRing = newRing
	return node
}

// GetNode 根据key获取对应的节点
func (ch *Hash[A, T]) GetNode(key string) T {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 {
		var zero T
		return zero
	}

	// 计算key的哈希值
	hash := hashKey(key)

	// 二分查找第一个大于等于hash的槽位
	idx := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i].hash >= hash
	})

	// 如果没找到，返回第一个（环形）
	if idx == len(ch.hashRing) {
		idx = 0
	}

	return ch.hashRing[idx].node
}

// GetNodeByUint64 根据uint64类型的key获取节点
func (ch *Hash[A, T]) GetNodeByHash(key uint64) T {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 {
		var zero T
		return zero
	}

	// 计算哈希值
	hash := hashUint64(key)

	// 二分查找
	idx := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i].hash >= hash
	})

	if idx == len(ch.hashRing) {
		idx = 0
	}

	return ch.hashRing[idx].node
}

// GetNodeByID 根据节点 ID 获取节点
func (ch *Hash[A, T]) GetNodeByKey(nodeID A) T {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.nodes[nodeID]
}

// GetNodes 获取所有节点
func (ch *Hash[A, T]) GetNodes() []T {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	nodes := make([]T, 0, len(ch.nodes))
	for _, node := range ch.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodeCount 获取节点数量
func (ch *Hash[A, T]) GetNodeCount() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.nodes)
}

// GetVirtualNodeCount 获取虚拟节点数量（即哈希环槽位数）
func (ch *Hash[A, T]) GetVirtualNodeCount() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.hashRing)
}

// GetNodesForKey 获取某个key的所有备份节点（用于故障转移）
func (ch *Hash[A, T]) GetNodesByKey(key string, count int) []T {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 || count <= 0 {
		return nil
	}

	hash := hashKey(key)
	// 起点 = 首个 hash >= key 的槽位；越界即回绕（环形）
	base := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i].hash >= hash
	}) % len(ch.hashRing)

	nodes := make([]T, 0, count)
	seen := make(map[A]bool)

	// 沿环顺时针取前 count 个不同真实节点（按槽位所属节点 id 判重）
	for i := 0; i < len(ch.hashRing) && len(nodes) < count; i++ {
		s := ch.hashRing[(base+i)%len(ch.hashRing)]
		if !seen[s.a] {
			seen[s.a] = true
			nodes = append(nodes, s.node)
		}
	}
	return nodes
}

// Clear 清空所有节点
func (ch *Hash[A, T]) Clear() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.nodes = make(map[A]T)
	ch.hashRing = make([]ringSlot[A, T], 0)
}

// UpdateNode 更新节点信息（按节点 id；环上该节点的槽位引用随之刷新、位置不变）
func (ch *Hash[A, T]) UpdateNode(nodeID A, node T) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否存在
	if _, exists := ch.nodes[nodeID]; !exists {
		return fmt.Errorf("node not found: %v", nodeID)
	}

	// 更新真实节点
	ch.nodes[nodeID] = node

	// C：刷新环上属于该节点的槽位引用（保持位置不变）
	for i := range ch.hashRing {
		if ch.hashRing[i].a == nodeID {
			ch.hashRing[i].node = node
		}
	}
	return nil
}
