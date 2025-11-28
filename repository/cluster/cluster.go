package cluster

import (
	"framework/define"
	"framework/internal/cluster"
	"framework/library/yaml"
	"framework/repository/cluster/internal/entity"
	"framework/repository/cluster/internal/service"
)

var (
	serviceObj = service.NewClusterService(define.MAX_NODE_TYPE_COUNT, entity.NewCluster)
)

func init() {
	cluster.Set(serviceObj.Get)
}

func Init(cfg *yaml.EtcdConfig) error {
	return serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}
