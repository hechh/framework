package domain

import (
	"framework/packet"
)

const (
	CLUSTER_BUCKET_SIZE = 256 // 集群桶的数量
	ETCD_GRANT_TTL      = 15
)

// 服务发现接口
type IWatcher interface {
	Watch(func(string, []byte)) error // 监听k-v变更
	Close()                           // 关闭监听服务
}

// 服务注册接口
type IRegister interface {
	Register(string, []byte) error // 注册服务节点
	Close()                        // 关闭服务注册服务
}

// 集群接口
type ICluster interface {
	Size() int                       // 集群节点数量
	Add(node *packet.Node)           // 添加节点
	Get(nodeId uint32) *packet.Node  // 获取节点
	Del(nodeId uint32) *packet.Node  // 删除节点
	Random(seed uint64) *packet.Node // 路由一个节点
}
