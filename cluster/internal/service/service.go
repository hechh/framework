package service

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/hechh/framework"
	"github.com/hechh/framework/cluster/internal/entity"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/yaml"

	"github.com/spf13/cast"
)

type ClusterService struct {
	register framework.IRegister
	watcher  framework.IWatcher
	clusters map[uint32]framework.ICluster
}

func NewService(max uint32) *ClusterService {
	ret := &ClusterService{clusters: make(map[uint32]framework.ICluster)}
	for i := uint32(1); i <= max; i++ {
		ret.clusters[i] = entity.NewCluster(i)
	}
	return ret
}

func (d *ClusterService) Init(cfg *yaml.EtcdConfig) (err error) {
	// 服务注册
	if d.register, err = entity.NewEtcdRegister(cfg.Topic, cfg.Endpoints); err != nil {
		return
	}

	nn := framework.GetSelf()
	buf, err := json.Marshal(nn)
	if err != nil {
		return err
	}
	if err := d.register.Register(fmt.Sprintf("%d/%d", nn.Type, nn.Id), buf); err != nil {
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

func (d *ClusterService) Get(nodeType uint32) framework.ICluster {
	if cls, ok := d.clusters[nodeType]; ok {
		return cls
	}
	return nil
}

func (d *ClusterService) addKeyValue(key string, value []byte) {
	nodeType := cast.ToUint32(cast.ToUint32(path.Base(path.Dir(key))))
	nodeId := cast.ToUint32(path.Base(key))
	// 获取集群
	cluster := d.Get(nodeType)
	if cluster == nil {
		mlog.Errorf("节点类型(%d)不支持", nodeType)
		return
	}
	// 删除节点？
	if value == nil {
		nn := cluster.Del(nodeId)
		mlog.Infof("删除服务节点:%d/%d  %s:%d", nn.GetType(), nn.GetId(), nn.GetIp(), nn.GetPort())
		return
	}
	// 添加节点
	nn := &packet.Node{}
	if err := json.Unmarshal(value, nn); err != nil {
		mlog.Errorf("节点解析错误:%v", err)
		return
	}
	cluster.Add(nn)
	mlog.Infof("添加服务节点:%d/%d  %s:%d", nn.GetType(), nn.GetId(), nn.GetIp(), nn.GetPort())
}
