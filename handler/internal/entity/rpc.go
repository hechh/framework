package entity

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/hechh/framework"
)

type Rpc[T any, U any] struct {
	*Base
	nodeType uint32
	cmd      uint32
}

func NewRpcHandler[T any, U any](en framework.ISerialize, nodeType uint32, cmd uint32, name string) *Rpc[T, U] {
	return &Rpc[T, U]{
		Base:     NewBase(en, name, reflect.ValueOf(nil)),
		nodeType: nodeType,
		cmd:      cmd,
	}
}

func (d *Rpc[T, U]) GetNodeType() uint32 {
	return d.nodeType
}

func (d *Rpc[T, U]) GetCmd() uint32 {
	return d.cmd
}

func (d *Rpc[T, U]) NewReq() proto.Message {
	return any(new(T)).(proto.Message)
}

func (d *Rpc[T, U]) NewRsp() proto.Message {
	return any(new(U)).(proto.Message)
}
