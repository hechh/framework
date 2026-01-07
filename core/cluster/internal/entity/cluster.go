package entity

import (
	"framework/core"
	"framework/library/pool"
	"framework/library/random"
	"framework/packet"
	"math"
	"sync"
)

type Cluster struct {
	mutex    sync.RWMutex
	nodeType uint32
	data     map[uint32]*packet.Node
	buckets  [core.CLUSTER_BUCKET_SIZE]*packet.Node
}

func NewCluster(nodeType uint32) core.ICluster {
	return &Cluster{
		nodeType: nodeType,
		data:     make(map[uint32]*packet.Node),
	}
}

// 当前节点数量
func (c *Cluster) Size() int {
	return len(c.data)
}

// 获取节点
func (c *Cluster) Get(nodeId uint32) *packet.Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.data[nodeId]; ok {
		return val
	}
	return nil
}

// 删除节点
func (c *Cluster) Del(nodeId uint32) (nn *packet.Node) {
	if nn = c.Get(nodeId); nn != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		ll := len(c.data)
		newVal := math.Ceil(core.CLUSTER_BUCKET_SIZE / float64(ll-1))
		oldVal := math.Ceil(core.CLUSTER_BUCKET_SIZE / float64(ll))
		diff := int(newVal - oldVal)

		// 删除buckets中的节点
		pos := 0
		for _, item := range c.data {
			count := 0
			for ; pos < core.CLUSTER_BUCKET_SIZE; pos++ {
				if c.buckets[pos].Id == nodeId {
					c.buckets[pos] = item
					count++
					if count == diff {
						break
					}
				}
			}
		}
		delete(c.data, nodeId)
	}
	return
}

func (c *Cluster) Add(node *packet.Node) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	ll := len(c.data)
	newVal := math.Ceil(core.CLUSTER_BUCKET_SIZE / float64(ll))
	oldVal := math.Ceil(core.CLUSTER_BUCKET_SIZE / float64(ll+1))
	diff := int(newVal - oldVal)

	// 在buckets中添加
	tmps := map[uint32]int{}
	for pos := 0; pos < core.CLUSTER_BUCKET_SIZE; pos++ {
		item := c.buckets[pos]
		val, ok := tmps[item.Id]
		if ok && val == diff {
			continue
		}
		tmps[item.Id]++
		c.buckets[pos] = node
	}
	c.data[node.Id] = node
}

func (c *Cluster) Random(seed uint64) *packet.Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if seed <= 0 {
		return c.buckets[random.Intn(core.CLUSTER_BUCKET_SIZE)]
	}
	// hash路由
	h := pool.GetHash64()
	defer pool.PutHash64(h)
	h.Reset()
	var b [8]byte
	b[0] = byte(seed >> 56)
	b[1] = byte(seed >> 48)
	b[2] = byte(seed >> 40)
	b[3] = byte(seed >> 32)
	b[4] = byte(seed >> 24)
	b[5] = byte(seed >> 16)
	b[6] = byte(seed >> 8)
	b[7] = byte(seed)
	h.Write(b[:])
	return c.buckets[h.Sum64()%core.CLUSTER_BUCKET_SIZE]
}
