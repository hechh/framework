package cluster

import (
	"framework/core/cluster/internal/service"
	"framework/core/define"
	"framework/library/yaml"
)

var (
	serviceObj = service.NewService(define.MAX_NODE_TYPE_COUNT)
)

func Init(cfg *yaml.EtcdConfig) error {
	return serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func Get(nodeType uint32) define.ICluster {
	return serviceObj.Get(nodeType)
}
