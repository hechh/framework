package cluster

import (
	"framework/define"
	"framework/packet"
)

// 获取指定节点类型的集群
type GetFunc func(int32) define.ICluster

var (
	getFunc GetFunc
)

func Set(f GetFunc) {
	getFunc = f
}

func Get(nodeType int32) define.ICluster {
	return getFunc(nodeType)
}

func GetrNode(nodeType int32, nodeId int32) *packet.Node {
	if cls := Get(nodeType); cls != nil {
		return cls.Get(nodeId)
	}
	return nil
}
