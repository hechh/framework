package cluster

import (
	"encoding/json"
	"framework/define"
	"framework/library/mlog"
	"framework/packet"
	"path"

	"github.com/spf13/cast"
)

type ClusterMgr struct {
	clusters map[int32]*Cluster
}

func NewClusterMgr(max int32) *ClusterMgr {
	ret := &ClusterMgr{clusters: make(map[int32]*Cluster)}
	for i := int32(1); i <= max; i++ {
		ret.clusters[i] = NewCluster(i)
	}
	return ret
}

func (d *ClusterMgr) Get(nodeType int32) define.ICluster {
	if cls, ok := d.clusters[nodeType]; ok {
		return cls
	}
	return nil
}

func (d *ClusterMgr) AddKeyValue(key string, value []byte) {
	nodeType := cast.ToInt32(cast.ToUint32(path.Base(path.Dir(key))))
	nodeId := cast.ToInt32(path.Base(key))
	// 获取集群
	cluster := d.Get(nodeType)
	if cluster == nil {
		mlog.Error(0, "节点类型(%d)不支持", nodeType)
		return
	}
	// 删除节点？
	if value == nil {
		nn := cluster.Del(nodeId)
		mlog.Info(0, "删除服务节点:%d/%d  %s:%d", nn.GetType(), nn.GetId(), nn.GetIp(), nn.GetPort())
		return
	}
	// 添加节点
	nn := &packet.Node{}
	if err := json.Unmarshal(value, nn); err != nil {
		mlog.Error(0, "节点解析错误:%v", err)
		return
	}
	cluster.Add(nn)
	mlog.Info(0, "添加服务节点:%d/%d  %s:%d", nn.GetType(), nn.GetId(), nn.GetIp(), nn.GetPort())
}
