package base

import (
	"framework/internal/common"
	"reflect"
)

type Base struct {
	nodeType uint32
	name     string
	id       uint32
	cmd      uint32
}

func NewBase(nodeType uint32, cmd uint32, fun reflect.Value) *Base {
	name := common.ParseActorFunc(fun)
	return &Base{
		nodeType: nodeType,
		id:       common.GetCrc32(name),
		cmd:      cmd,
		name:     name,
	}
}

func (d *Base) GetType() uint32 {
	return d.nodeType
}

func (d *Base) GetId() uint32 {
	return d.id
}

func (d *Base) GetCmd() uint32 {
	return d.cmd
}

func (d *Base) GetName() string {
	return d.name
}
