package cluster

import (
	"framework/define"
	"framework/library/yaml"
	"framework/packet"
	"framework/repository/cluster/internal/service"
)

var (
	serviceObj = service.NewService(define.MAX_NODE_TYPE_COUNT)
)

func Init(cfg *yaml.EtcdConfig, nn *packet.Node) error {
	return serviceObj.Init(cfg, nn)
}

func Close() {
	serviceObj.Close()
}

func Get(nodeType uint32) define.ICluster {
	return serviceObj.Get(nodeType)
}
