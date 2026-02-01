package entity

import (
	"github.com/hechh/framework"
)

type Rpc[T any, U any] struct {
	framework.ISerialize
	name     string
	crc32    uint32
	nodeType uint32
	cmd      uint32
}

func NewRpc[T any, U any](en framework.ISerialize, nodeType uint32, cmd uint32, name string) *Rpc[T, U] {
	return &Rpc[T, U]{
		ISerialize: en,
		name:       name,
		crc32:      framework.GetCrc32(name),
		nodeType:   nodeType,
		cmd:        cmd,
	}
}

func (d *Rpc[T, U]) GetName() string {
	return d.name
}

func (d *Rpc[T, U]) GetCrc32() uint32 {
	return d.crc32
}

func (d *Rpc[T, U]) GetNodeType() uint32 {
	return d.nodeType
}

func (d *Rpc[T, U]) GetCmd() uint32 {
	return d.cmd
}

func (d *Rpc[T, U]) New(pos int) any {
	switch pos {
	case 0:
		return new(T)
	case 1:
		return new(U)
	}
	return nil
}
