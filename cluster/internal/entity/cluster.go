package entity

import (
	"math"
	"sync"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/random"
)

type Cluster struct {
	mutex    sync.RWMutex
	nodeType uint32
	data     map[uint32]*packet.Node
	buckets  [framework.CLUSTER_BUCKET_SIZE]uint32
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
		nodeId := c.buckets[pos]
		if nodeId <= 0 || tmps[nodeId] >= per {
			nodeId = node.Id
		}
		if tmps[nodeId] < per {
			tmps[nodeId]++
			c.buckets[pos] = nodeId
		}
	}
	mlog.Tracef("[cluster] 添加节点%v", c.buckets)
}

// 删除节点
func (c *Cluster) Del(nodeId uint32) (nn *packet.Node) {
	nn = c.Get(nodeId)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, nodeId)
	if len(c.data) <= 0 {
		for pos := 0; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
			c.buckets[pos] = 0
		}
		return
	}

	tmps := map[uint32]int{}
	for pos := 0; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
		if nodeId := c.buckets[pos]; nodeId != nodeId {
			tmps[nodeId]++
		}
	}

	pos := 0
	per := int(math.Ceil(framework.CLUSTER_BUCKET_SIZE / float64(len(c.data))))
	for _, node := range c.data {
		for ; pos < framework.CLUSTER_BUCKET_SIZE; pos++ {
			if nodeId := c.buckets[pos]; nodeId == nodeId {
				if tmps[node.Id] >= per {
					break
				}
				tmps[node.Id]++
				c.buckets[pos] = nodeId
			}
		}
	}
	mlog.Tracef("[cluster] 删除节点：%v", c.buckets)
	return
}

func (c *Cluster) Random(seed uint64) *packet.Node {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if seed <= 0 {
		if nodeId := c.buckets[random.Intn(framework.CLUSTER_BUCKET_SIZE)]; nodeId > 0 {
			return c.data[nodeId]
		}
		return nil
	}
	if nodeId := c.buckets[seed%framework.CLUSTER_BUCKET_SIZE]; nodeId > 0 {
		return c.data[nodeId]
	}
	return nil
	/*
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
		nodeId := c.buckets[h.Sum64()%framework.CLUSTER_BUCKET_SIZE]
		if nodeId > 0 {
			return c.data[nodeId]
		}
		return nil
	*/
}
