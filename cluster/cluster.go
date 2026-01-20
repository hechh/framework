package cluster

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/cluster/internal/service"
	"github.com/hechh/library/yaml"
)

var (
	serviceObj = service.NewService(framework.MAX_NODE_TYPE_COUNT)
)

func init() {
	framework.SetCluster(Get)
}

func Init(cfg *yaml.EtcdConfig) error {
	return serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func Get(nodeType uint32) framework.ICluster {
	return serviceObj.Get(nodeType)
}
