package config

import "github.com/hechh/framework/packet"

var (
	object *Config
)

func SetObject(obj *Config) {
	object = obj
}

func Get() *Config {
	return object
}

func GetCommonCfg() *CommonConfig {
	return object.Common
}

func GetSelfConfig() *NodeConfig {
	return object.node
}

func GetGatewayConfig() *NodeConfig {
	return object.node
}

func GetSelfNode() *packet.Node {
	return object.self
}

func GetGatewayNodeType() uint32 {
	if object != nil && object.gateway != nil {
		return object.gateway.Type
	}
	return 0
}

func GetNodeConfig(nodeType, nodeId uint32) *NodeConfig {
	if items, ok := object.types[nodeType]; ok {
		return items[nodeId]
	}
	return nil
}
