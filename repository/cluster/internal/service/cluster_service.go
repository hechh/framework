package service

import (
	"encoding/json"
	"fmt"
	"framework/define"
	"framework/internal/global"
	"framework/library/mlog"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/cluster/internal/entity"
	"path"

	"github.com/spf13/cast"
)

type ClusterService struct {
	register define.IRegister
	watcher  define.IWatcher
	clusters map[int32]define.ICluster
}

func NewClusterService(max int32, f func(int32) define.ICluster) *ClusterService {
	ret := &ClusterService{clusters: make(map[int32]define.ICluster)}
	for i := int32(1); i <= max; i++ {
		ret.clusters[i] = f(i)
	}
	return ret
}

func (d *ClusterService) Init(cfg *yaml.EtcdConfig) (err error) {
	// 服务注册
	if d.register, err = entity.NewEtcdRegister(cfg.Topic, cfg.Endpoints); err != nil {
		return
	}

	self := global.GetSelf()
	buf, err := json.Marshal(self)
	if err != nil {
		return err
	}
	if err := d.register.Register(fmt.Sprintf("%d/%d", self.Type, self.Id), buf); err != nil {
		return err
	}

	// 服务监听
	if d.watcher, err = entity.NewEtcdWatcher(cfg.Topic, cfg.Endpoints); err != nil {
		return
	}
	if err := d.watcher.Watch(d.addKeyValue); err != nil {
		return err
	}
	return
}

func (d *ClusterService) Close() {
	d.watcher.Close()
	d.register.Close()
}
func (d *ClusterService) Get(nodeType int32) define.ICluster {
	if cls, ok := d.clusters[nodeType]; ok {
		return cls
	}
	return nil
}

func (d *ClusterService) addKeyValue(key string, value []byte) {
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
