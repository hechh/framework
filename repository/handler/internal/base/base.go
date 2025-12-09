package base

import (
	"hash/crc32"
	"reflect"
	"runtime"
	"strings"
)

type Base struct {
	nodeType uint32
	name     string
	id       uint32
	cmd      uint32
}

func NewBase(nodeType uint32, cmd uint32, fun reflect.Value) *Base {
	runName := runtime.FuncForPC(fun.Pointer()).Name()
	strs := strings.Split(runName, "(*")
	name := strings.ReplaceAll(strs[len(strs)-1], ")", "")
	return &Base{
		nodeType: nodeType,
		id:       crc32.ChecksumIEEE([]byte(name)),
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
