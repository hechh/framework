package consistent

import (
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"github.com/hechh/framework/packet"
)

// ConsistentHash 一致性哈希实现
type ConsistentHash struct {
	mu       sync.RWMutex
	nodes    map[uint32]*packet.Node // 真实节点
	virtual  map[uint32]*packet.Node // 虚拟节点
	hashRing []uint32                // 哈希环（已排序）
	virtuals int                     // 每个真实节点的虚拟节点数
}

// NewConsistentHash 创建一致性哈希实例
func NewConsistentHash(virtuals int) *ConsistentHash {
	if virtuals <= 0 {
		virtuals = 150 // 默认虚拟节点数
	}
	return &ConsistentHash{
		nodes:    make(map[uint32]*packet.Node),
		virtual:  make(map[uint32]*packet.Node),
		hashRing: make([]uint32, 0),
		virtuals: virtuals,
	}
}

// AddNode 添加节点
func (ch *ConsistentHash) AddNode(node *packet.Node) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否已存在
	if _, exists := ch.nodes[node.Id]; exists {
		return fmt.Errorf("node already exists: %v", node)
	}

	// 添加真实节点
	ch.nodes[node.Id] = node

	// 添加虚拟节点
	for i := 0; i < ch.virtuals; i++ {
		virtualKey := ch.hashVirtualNode(node.Id, i)
		ch.virtual[virtualKey] = node
		ch.hashRing = append(ch.hashRing, virtualKey)
	}

	// 重新排序哈希环
	sort.Slice(ch.hashRing, func(i, j int) bool {
		return ch.hashRing[i] < ch.hashRing[j]
	})
	return nil
}

// RemoveNode 移除节点
func (ch *ConsistentHash) RemoveNode(nodeID uint32) *packet.Node {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否存在
	node, exists := ch.nodes[nodeID]
	if !exists {
		return nil
	}

	// 移除真实节点
	delete(ch.nodes, nodeID)

	// 移除虚拟节点
	newRing := make([]uint32, 0, len(ch.hashRing))
	for _, hash := range ch.hashRing {
		if n, ok := ch.virtual[hash]; ok && n.Id == nodeID {
			delete(ch.virtual, hash)
		} else {
			newRing = append(newRing, hash)
		}
	}
	ch.hashRing = newRing
	return node
}

// GetNode 根据key获取对应的节点
func (ch *ConsistentHash) GetNode(key string) *packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 {
		return nil
	}

	// 计算key的哈希值
	hash := ch.hashKey(key)

	// 二分查找第一个大于等于hash的节点
	idx := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i] >= hash
	})

	// 如果没找到，返回第一个（环形）
	if idx == len(ch.hashRing) {
		idx = 0
	}

	virtualKey := ch.hashRing[idx]
	return ch.virtual[virtualKey]
}

// GetNodeByUint64 根据uint64类型的key获取节点
func (ch *ConsistentHash) GetNodeByUint64(key uint64) *packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 {
		return nil
	}

	// 计算哈希值
	hash := ch.hashUint64(key)

	// 二分查找
	idx := sort.Search(len(ch.hashRing), func(i int) bool {
		return ch.hashRing[i] >= hash
	})

	if idx == len(ch.hashRing) {
		idx = 0
	}

	virtualKey := ch.hashRing[idx]
	return ch.virtual[virtualKey]
}

// GetNodeByID 根据节点ID获取节点
func (ch *ConsistentHash) GetNodeByID(nodeID uint32) *packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.nodes[nodeID]
}

// GetNodes 获取所有节点
func (ch *ConsistentHash) GetNodes() []*packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	nodes := make([]*packet.Node, 0, len(ch.nodes))
	for _, node := range ch.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetNodesByType 根据类型获取节点列表
func (ch *ConsistentHash) GetNodesByType(nodeType uint32) []*packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	nodes := make([]*packet.Node, 0)
	for _, node := range ch.nodes {
		if node.Type == nodeType {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

// GetNodeCount 获取节点数量
func (ch *ConsistentHash) GetNodeCount() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.nodes)
}

// GetVirtualNodeCount 获取虚拟节点数量
func (ch *ConsistentHash) GetVirtualNodeCount() int {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return len(ch.virtual)
}

// hashKey 计算字符串key的哈希值（使用FNV-32a避免内存分配）
func (ch *ConsistentHash) hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

// hashUint64 计算uint64 key的哈希值
func (ch *ConsistentHash) hashUint64(key uint64) uint32 {
	h := fnv.New32a()
	// 使用高效的二进制写入，避免make和binary.BigEndian
	var buf [8]byte
	buf[0] = byte(key >> 56)
	buf[1] = byte(key >> 48)
	buf[2] = byte(key >> 40)
	buf[3] = byte(key >> 32)
	buf[4] = byte(key >> 24)
	buf[5] = byte(key >> 16)
	buf[6] = byte(key >> 8)
	buf[7] = byte(key)
	h.Write(buf[:])
	return h.Sum32()
}

// hashVirtualNode 计算虚拟节点的哈希值
func (ch *ConsistentHash) hashVirtualNode(nodeID uint32, index int) uint32 {
	h := fnv.New32a()
	// 手动格式化字符串避免 fmt.Sprintf 的内存分配
	// 格式: "nodeID:index"，nodeID 最大 10位，index 最大 10位，加冒号 1位，共 21 字节
	var buf [21]byte
	n := 0

	// 写入 nodeID
	if nodeID == 0 {
		buf[n] = '0'
		n++
	} else {
		temp := [10]byte{}
		i := 10
		t := nodeID
		for t > 0 {
			i--
			temp[i] = byte('0' + t%10)
			t /= 10
		}
		copy(buf[n:], temp[i:])
		n += 10 - i
	}

	buf[n] = ':'
	n++

	// 写入 index
	if index == 0 {
		buf[n] = '0'
		n++
	} else {
		temp := [10]byte{}
		i := 10
		t := index
		neg := false
		if t < 0 {
			neg = true
			t = -t
		}
		for t > 0 {
			i--
			temp[i] = byte('0' + t%10)
			t /= 10
		}
		if neg {
			i--
			temp[i] = '-'
		}
		copy(buf[n:], temp[i:])
		n += 10 - i
	}

	h.Write(buf[:n])
	return h.Sum32()
}

// GetNodesForKey 获取某个key的所有备份节点（用于故障转移）
func (ch *ConsistentHash) GetNodesForKey(key string, count int) []*packet.Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	if len(ch.hashRing) == 0 || count <= 0 {
		return nil
	}

	hash := ch.hashKey(key)
	nodes := make([]*packet.Node, 0, count)
	seen := make(map[uint32]bool)

	// 获取前count个不同真实节点
	for i := 0; i < len(ch.hashRing) && len(nodes) < count; i++ {
		// 计算环上的位置
		pos := (sort.Search(len(ch.hashRing), func(j int) bool {
			return ch.hashRing[j] >= hash
		}) + i) % len(ch.hashRing)

		virtualKey := ch.hashRing[pos]
		node := ch.virtual[virtualKey]

		// 跳过已选择的节点
		if !seen[node.Id] {
			seen[node.Id] = true
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// Clear 清空所有节点
func (ch *ConsistentHash) Clear() {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	ch.nodes = make(map[uint32]*packet.Node)
	ch.virtual = make(map[uint32]*packet.Node)
	ch.hashRing = make([]uint32, 0)
}

// UpdateNode 更新节点信息
func (ch *ConsistentHash) UpdateNode(node *packet.Node) error {
	if node == nil {
		return fmt.Errorf("node cannot be nil")
	}

	ch.mu.Lock()
	defer ch.mu.Unlock()

	// 检查节点是否存在
	if _, exists := ch.nodes[node.Id]; !exists {
		return fmt.Errorf("node not found: %d", node.Id)
	}

	// 更新真实节点
	ch.nodes[node.Id] = node

	// 更新所有虚拟节点指向
	for _, virtualKey := range ch.hashRing {
		if n, ok := ch.virtual[virtualKey]; ok && n.Id == node.Id {
			ch.virtual[virtualKey] = node
		}
	}

	return nil
}
