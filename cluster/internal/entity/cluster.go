package entity

import (
	"math"
	"sync"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/pool"
	"github.com/hechh/library/random"
)

type Cluster struct {
	mutex    sync.RWMutex
	nodeType uint32
	data     map[uint32]*packet.Node
	buckets  [framework.CLUSTER_BUCKET_SIZE]*packet.Node
}

func NewCluster(nodeType uint32) framework.ICluster {
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

func (c *Cluster) Add(node *packet.Node) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[node.Id] = node
	per := int(math.Ceil(framework.CLUSTER_BUCKET_SIZE / float64(len(c.data))))
	tmps := map[uint32]int{}
	for pos := 0; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
		item := c.buckets[pos]
		if item == nil || tmps[item.Id] >= per {
			item = node
		}
		if tmps[item.Id] < per {
			tmps[item.Id]++
			c.buckets[pos] = node
		}
	}
}

// 删除节点
func (c *Cluster) Del(nodeId uint32) (nn *packet.Node) {
	nn = c.Get(nodeId)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, nodeId)
	if len(c.data) <= 0 {
		for pos := 0; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
			c.buckets[pos] = nil
		}
		return
	}

	tmps := map[uint32]int{}
	for pos := 0; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
		if item := c.buckets[pos]; item.Id != nodeId {
			tmps[item.Id]++
		}
	}

	pos := 0
	per := int(math.Ceil(framework.CLUSTER_BUCKET_SIZE / float64(len(c.data))))
	for _, node := range c.data {
		for ; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
			if item := c.buckets[pos]; item.Id == nodeId {
				if tmps[node.Id] >= per {
					break
				}
				tmps[node.Id]++
				c.buckets[pos] = node
			}
		}
	}
	return
}

func (c *Cluster) Random(seed uint64) *packet.Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if seed <= 0 {
		return c.buckets[random.Intn(framework.CLUSTER_BUCKET_SIZE)]
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
	return c.buckets[h.Sum64()%framework.CLUSTER_BUCKET_SIZE]
}
