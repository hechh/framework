package entity

import (
	"reflect"

	"github.com/hechh/framework"
)

type Rpc1[T any] struct {
	*Base
	nodeType uint32
	cmd      uint32
}

func NewRpc1Handler[T any](en framework.ISerialize, nodeType uint32, cmd uint32, name string) *Rpc1[T] {
	return &Rpc1[T]{
		Base:     NewBase(en, name, reflect.ValueOf(nil)),
		nodeType: nodeType,
		cmd:      cmd,
	}
}

func (d *Rpc1[T]) GetNodeType() uint32 {
	return d.nodeType
}

func (d *Rpc1[T]) GetCmd() uint32 {
	return d.cmd
}

func (d *Rpc1[T]) New(pos int) any {
	switch pos {
	case 0:
		return new(T)
	}
	return nil
}

type Rpc2[T any, U any] struct {
	*Base
	nodeType uint32
	cmd      uint32
}

func NewRpc2Handler[T any, U any](en framework.ISerialize, nodeType uint32, cmd uint32, name string) *Rpc2[T, U] {
	return &Rpc2[T, U]{
		Base:     NewBase(en, name, reflect.ValueOf(nil)),
		nodeType: nodeType,
		cmd:      cmd,
	}
}

func (d *Rpc2[T, U]) GetNodeType() uint32 {
	return d.nodeType
}

func (d *Rpc2[T, U]) GetCmd() uint32 {
	return d.cmd
}

func (d *Rpc2[T, U]) New(pos int) any {
	switch pos {
	case 0:
		return new(T)
	case 1:
		return new(U)
	}
	return nil
}
