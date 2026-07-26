package router

import (
	"sync/atomic"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
)

type Entity struct {
	idType     uint32                    // 实体类型
	id         uint64                    // 实体ID
	nodes      map[uint32]*atomic.Uint32 // 节点类型到路由ID的映射
	change     atomic.Uint32             // 路由映射是否变更
	accessTime atomic.Int64              // 最后访问时间
}

// NewEntity 创建路由实体
func NewEntity(idType uint32, id uint64, nodeTypes map[uint32]struct{}) *Entity {
	ret := &Entity{
		idType: idType,
		id:     id,
		nodes:  make(map[uint32]*atomic.Uint32, len(nodeTypes)),
	}
	for nodeType := range nodeTypes {
		ret.nodes[nodeType] = new(atomic.Uint32)
	}
	ret.accessTime.Store(datetime.NowUnix())
	return ret
}

// GetAccessTime 获取最后访问时间
func (d *Entity) GetAccessTime() int64 {
	return d.accessTime.Load()
}

// UpdateAccessTime 更新最后访问时间
func (d *Entity) UpdateAccessTime() {
	d.accessTime.Store(datetime.NowUnix())
}

// IsChange 是否变更
func (d *Entity) IsChange() bool {
	return d.change.Load() > 0
}

// Reset 重置变更映射
func (d *Entity) Reset() {
	d.change.Store(0)
}

// Get 获取节点路由ID
func (d *Entity) Get(nodeType uint32) uint32 {
	if item, ok := d.nodes[nodeType]; ok {
		return item.Load()
	}
	return 0
}

// Set 设置节点路由ID
func (d *Entity) Set(nodeType uint32, nodeId uint32) {
	item := d.nodes[nodeType]
	if item.Load() != nodeId {
		item.Store(nodeId)
		d.change.Add(1)
	}
}

// GetNodes 获取所有节点路由ID
func (d *Entity) GetNodes() []*packet.RouteInfo {
	rets := make([]*packet.RouteInfo, 0, len(d.nodes))
	for nodeType, item := range d.nodes {
		if nodeId := item.Load(); nodeId > 0 {
			rets = append(rets, &packet.RouteInfo{
				Type: nodeType,
				Id:   nodeId,
			})
		}
	}
	return rets
}

// SetNodes 设置所有节点路由ID
func (d *Entity) SetNodes(items ...*packet.RouteInfo) {
	flag := false
	for _, n := range items {
		if n.Type == 0 || n.Id == 0 {
			continue
		}
		if item := d.nodes[n.Type]; item.Load() != n.Id {
			item.Store(n.Id)
			flag = true
		}
	}
	if flag {
		d.change.Add(1)
	}
}

// ToRouteInfo 转换为路由信息
func (d *Entity) ToRouteInfo() *packet.Router {
	return &packet.Router{
		IdType: d.idType,
		Id:     d.id,
		List:   d.GetNodes(),
	}
}
