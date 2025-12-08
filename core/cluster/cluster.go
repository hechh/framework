package cluster

import (
	"framework/core/cluster/domain"
	"framework/core/cluster/internal/service"
	"framework/core/define"
	"framework/library/yaml"
	"framework/packet"
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

func Get(nodeType uint32) domain.ICluster {
	return serviceObj.Get(nodeType)
}
