package cluster

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/core/cluster/internal/consistent"
	"github.com/hechh/framework/core/cluster/internal/domain"
	"github.com/hechh/library/mlog"
	"google.golang.org/protobuf/proto"
)

type Cluster struct {
	disc     domain.IDiscovery                     // 服务发现接口
	virtuals map[uint32]*consistent.ConsistentHash // 节点类型到一致性哈希的映射
}

// NewCluster 创建集群实例
func NewCluster(dis domain.IDiscovery) *Cluster {
	return &Cluster{
		disc:     dis,
		virtuals: make(map[uint32]*consistent.ConsistentHash),
	}
}

// Init 初始化集群服务（实现 IComponent 接口）
func (d *Cluster) Init(cfg *packet.Config) error {
	discoveryCfg := cfg.Discovery
	nodeCfg := cfg.Node
	self := &packet.Node{
		Type: nodeCfg.Type,
		Id:   nodeCfg.Id,
		Name: nodeCfg.Name,
		Ip:   nodeCfg.Ip,
		Port: int32(nodeCfg.Port),
	}
	nodeTypes := make(map[uint32]struct{}, len(cfg.Types))
	for t := range cfg.Types {
		nodeTypes[t] = struct{}{}
	}

	// 初始化服务发现接口
	if err := d.disc.Init(discoveryCfg); err != nil {
		return err
	}

	// 初始化各节点类型的一致性哈希
	for nodeType := range nodeTypes {
		d.virtuals[nodeType] = consistent.NewConsistentHash(domain.DEFAULT_VIRTUAL_COUNT)
	}

	// 序列化节点信息并注册
	data, err := proto.Marshal(self)
	if err != nil {
		d.disc.Close()
		return fmt.Errorf("marshal failed: %w", err)
	}

	// 注册节点
	key := fmt.Sprintf("%d/%d", self.Type, self.Id)
	if err = d.disc.Register(key, data); err != nil {
		mlog.Errorf("注册节点失败 node:%v, error:%v", self, err)
		d.disc.Close()
		return err
	}

	// 启动服务发现监听
	err = d.disc.Watch(func(key string, val []byte) {
		if val == nil {
			d.handleNodeOffline(key)
		} else if len(val) > 0 {
			d.handleNodeOnline(val)
		}
	})
	if err != nil {
		mlog.Errorf("启动服务发现监听失败: %v", err)
		d.disc.Close()
	} else {
		mlog.Infof("集群初始化成功")
	}
	return nil
}

// Close 关闭集群，按顺序：register → watcher → shared client
func (d *Cluster) Close() {
	d.disc.Close()
	mlog.Infof("集群已关闭")
}

// handleNodeOnline 处理节点上线
func (d *Cluster) handleNodeOnline(data []byte) {
	node := &packet.Node{}
	if err := proto.Unmarshal(data, node); err != nil {
		mlog.Errorf("解析节点信息失败:%v", err)
		return
	}

	ch, ok := d.virtuals[node.Type]
	if !ok {
		mlog.Errorf("节点类型不支持:%d", node.Type)
		return
	}

	if existing := ch.GetNodeByID(node.Id); existing != nil {
		if err := ch.UpdateNode(node); err != nil {
			mlog.Errorf("更新节点失败:%v", err)
			return
		}
	} else if err := ch.AddNode(node); err != nil {
		mlog.Errorf("添加节点失败:%v", err)
		return
	}
	mlog.Infof("节点上线: Type=%d, Id=%d, Name=%s, Addr=%s:%d", node.Type, node.Id, node.Name, node.Ip, node.Port)
}

// handleNodeOffline 处理节点下线
func (d *Cluster) handleNodeOffline(key string) {
	nodeType, nodeID := parseNodeKey(key)
	if nodeType == 0 || nodeID == 0 {
		return
	}

	ch, ok := d.virtuals[nodeType]
	if !ok {
		mlog.Errorf("节点类型不支持:%d", nodeType)
		return
	}

	node := ch.RemoveNode(nodeID)
	if node != nil {
		mlog.Infof("节点下线: Type=%d, Id=%d, Name=%s", node.Type, node.Id, node.Name)
	}
}

// parseNodeKey 从 etcd key 中解析节点类型和 ID
func parseNodeKey(key string) (uint32, uint32) {
	parts := strings.Split(key, "/")
	if len(parts) < 3 {
		return 0, 0
	}

	nodeTypeStr := parts[len(parts)-2]
	nodeIDStr := parts[len(parts)-1]

	nodeType, err1 := strconv.ParseUint(nodeTypeStr, 10, 32)
	nodeID, err2 := strconv.ParseUint(nodeIDStr, 10, 32)

	if err1 != nil || err2 != nil {
		return 0, 0
	}
	return uint32(nodeType), uint32(nodeID)
}

// Count 获取指定类型的节点数量
func (d *Cluster) Count(nodeType uint32) int {
	if ch, ok := d.virtuals[nodeType]; ok {
		return ch.GetNodeCount()
	}
	return 0
}

// Total 获取所有节点总数
func (d *Cluster) Total() int {
	total := 0
	for _, ch := range d.virtuals {
		total += ch.GetNodeCount()
	}
	return total
}

// Get 获取指定节点
func (d *Cluster) Get(nodeType, nodeId uint32) *packet.Node {
	cn, ok := d.virtuals[nodeType]
	if !ok {
		return nil
	}
	return cn.GetNodeByID(nodeId)
}

// Gets 根据类型获取所有节点
func (d *Cluster) Gets(nodeType uint32) []*packet.Node {
	if ch, ok := d.virtuals[nodeType]; ok {
		return ch.GetNodesByType(nodeType)
	}
	return nil
}

// Route 根据种子路由到指定节点（一致性哈希）
func (d *Cluster) Route(nodeType uint32, seed uint64) *packet.Node {
	if ch, ok := d.virtuals[nodeType]; ok {
		return ch.GetNodeByUint64(seed)
	}
	return nil
}

// Add 手动添加节点
func (d *Cluster) Add(node *packet.Node) {
	ch, ok := d.virtuals[node.Type]
	if !ok {
		mlog.Errorf("节点类型不支持:%v", node)
		return
	}
	if err := ch.AddNode(node); err != nil {
		mlog.Errorf("手动添加节点失败:%v", err)
	}
}

// Del 删除节点
func (d *Cluster) Del(nodeType, nodeID uint32) *packet.Node {
	ch, ok := d.virtuals[nodeType]
	if !ok {
		mlog.Errorf("节点类型不支持 nodeType:%d, nodeId:%d", nodeType, nodeID)
		return nil
	}

	if node := ch.RemoveNode(nodeID); node != nil {
		mlog.Infof("手动删除节点: Type=%d, Id=%d, Name=%s", node.Type, node.Id, node.Name)
		return node
	}
	return nil
}
