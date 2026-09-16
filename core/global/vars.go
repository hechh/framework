package global

import (
	"github.com/hechh/framework/library/enum"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/packet"
)

var (
	cmdConvertor  enum.IConvertor
	nodeConvertor enum.IConvertor
	gateway       enum.IEnum
	self          *packet.Node
	nodeTypes     []int32
)

func SetCmdConvertor(values map[int32]string, names map[string]int32) {
	cmdConvertor = enum.WrapConvertor(names, values)
}

func SetNodeConvertor(values map[int32]string, names map[string]int32) {
	nodeConvertor = enum.WrapConvertor(names, values)
	nodeTypes = tplutil.Map2Keys(values)
}

func SetSelf(n *packet.Node, gate enum.IEnum) {
	gateway = gate
	self = n
}

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

func HasCmd(val uint32) bool {
	return cmdConvertor.Has(val)
}
