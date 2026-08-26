package global

import (
	"github.com/hechh/framework/library/enum"
	"github.com/hechh/framework/packet"
)

var (
	cmdConvertor  enum.IConvertor
	nodeConvertor enum.IConvertor
	gateway       enum.IEnum
	self          *packet.Node
	nodeTypes     []int32
)

func GetSelf() *packet.Node {
	return self
}

func GetSelfName() string {
	return self.Name
}

func GetSelfNodeType() uint32 {
	return self.Type
}

func GetSelfNodeId() uint32 {
	return self.Id
}

func GetGatewayNodeType() uint32 {
	return uint32(gateway.Number())
}

func GetSupportNodeTypes() []int32 {
	return nodeTypes
}
