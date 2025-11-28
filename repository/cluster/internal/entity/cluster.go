package entity

import (
	"framework/define"
	"framework/library/pool"
	"framework/library/random"
	"framework/packet"
	"math"
	"sync"
)

type Cluster struct {
	mutex    sync.RWMutex
	nodeType int32
	data     map[int32]*packet.Node
	buckets  [define.CLUSTER_BUCKET_SIZE]*packet.Node
}

func NewCluster(nodeType int32) define.ICluster {
	return &Cluster{
		nodeType: nodeType,
		data:     make(map[int32]*packet.Node),
	}
}

// 当前节点数量
func (c *Cluster) Size() int {
	return len(c.data)
}

// 获取节点
func (c *Cluster) Get(nodeId int32) *packet.Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if val, ok := c.data[nodeId]; ok {
		return val
	}
	return nil
}

// 删除节点
func (c *Cluster) Del(nodeId int32) (nn *packet.Node) {
	if nn = c.Get(nodeId); nn != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		ll := len(c.data)
		newVal := math.Ceil(define.CLUSTER_BUCKET_SIZE / float64(ll-1))
		oldVal := math.Ceil(define.CLUSTER_BUCKET_SIZE / float64(ll))
		diff := int(newVal - oldVal)

		// 删除buckets中的节点
		pos := 0
		for _, item := range c.data {
			count := 0
			for ; pos < define.CLUSTER_BUCKET_SIZE; pos++ {
				if c.buckets[pos].GetId() == nodeId {
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
	newVal := math.Ceil(define.CLUSTER_BUCKET_SIZE / float64(ll))
	oldVal := math.Ceil(define.CLUSTER_BUCKET_SIZE / float64(ll+1))
	diff := int(newVal - oldVal)

	// 在buckets中添加
	tmps := map[int32]int{}
	for pos := 0; pos < define.CLUSTER_BUCKET_SIZE; pos++ {
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
		return c.buckets[random.Intn(define.CLUSTER_BUCKET_SIZE)]
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
	return c.buckets[h.Sum64()%define.CLUSTER_BUCKET_SIZE]
}
