package cluster

import (
	"framework/define"
	"framework/internal/cluster"
	"framework/library/yaml"
	"framework/repository/cluster/internal/entity"
	"framework/repository/cluster/internal/service"
)

var (
	obj *service.ClusterService
)

func Init(cfg *yaml.EtcdConfig) error {
	obj = service.NewClusterService(define.MAX_NODE_TYPE_COUNT, entity.NewCluster)
	cluster.Set(obj.Get)
	return obj.Init(cfg)
}

func Close() {
	if obj != nil {

		obj.Close()
	}
}
