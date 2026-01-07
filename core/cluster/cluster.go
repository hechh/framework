package cluster

import (
	"framework/core"
	"framework/core/cluster/internal/service"
	"framework/library/yaml"
)

var (
	serviceObj = service.NewService(core.MAX_NODE_TYPE_COUNT)
)

func init() {
	core.SetGetCluster(Get)
}

func Init(cfg *yaml.EtcdConfig) error {
	return serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func Get(nodeType uint32) core.ICluster {
	return serviceObj.Get(nodeType)
}
