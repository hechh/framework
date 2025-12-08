package entity

import "framework/define"

type GobRpc struct {
	*Common
	GobCrypto
}

func NewGobRpc(nodeType int32, cmd int32, actorFunc string) *GobRpc {
	return &GobRpc{
		Common: NewCommon(nodeType, cmd, actorFunc),
	}
}

func (d *GobRpc) Call(obj any, ctx define.IContext, args ...any) func() {
	return func() {}
}

func (d *GobRpc) Rpc(obj any, ctx define.IContext, body []byte) func() {
	return func() {}
}

type ProtoRpc struct {
	*Common
	ProtoCrypto
}

func NewProtoRpc(nodeType int32, cmd int32, actorFunc string) *ProtoRpc {
	return &ProtoRpc{
		Common: NewCommon(nodeType, cmd, actorFunc),
	}
}

func (d *ProtoRpc) Call(obj any, ctx define.IContext, args ...any) func() {
	return func() {}
}

func (d *ProtoRpc) Rpc(obj any, ctx define.IContext, body []byte) func() {
	return func() {}
}
