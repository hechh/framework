package handler

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/handler/internal/entity"
	"github.com/hechh/framework/handler/internal/service"
)

var (
	serviceObj = service.NewService()
)

func init() {
	framework.SetHandler(Get, GetCmdRpc, GetRpc)
}

func Name2Id(name string) (uint32, bool) {
	return serviceObj.Name2Id(name)
}

func Id2Name(val uint32) (string, bool) {
	return serviceObj.Id2Name(val)
}

func Get(name any) framework.IHandler {
	return serviceObj.Get(name)
}

func GetCmdRpc(cmd uint32) framework.IRpc {
	return serviceObj.GetCmdRpc(cmd)
}

func GetRpc(nodeType uint32, id any) framework.IRpc {
	return serviceObj.GetRpc(nodeType, id)
}

func RegisterRpc2[T any, U any](e framework.ISerialize, nodeType, cmd framework.IEnum, name string) {
	serviceObj.RegisterRpc(entity.NewRpc2Handler[T, U](e, nodeType.Integer(), cmd.Integer(), name))
}

func RegisterRpc1[T any](e framework.ISerialize, nodeType, cmd framework.IEnum, name string) {
	serviceObj.RegisterRpc(entity.NewRpc1Handler[T](e, nodeType.Integer(), cmd.Integer(), name))
}

func RegisterCmd[Actor any, V1 any, V2 any](f framework.P2Func[Actor, V1, V2]) {
	serviceObj.Register(entity.NewCmdHandler(framework.PROTO, f))
}

func Register0[Actor any](e framework.ISerialize, f framework.EmptyFunc[Actor]) {
	serviceObj.Register(entity.NewV0Handler(e, f))
}

func RegisterP1[Actor any, V1 any](e framework.ISerialize, f framework.P1Func[Actor, V1]) {
	serviceObj.Register(entity.NewP1Handler(e, f))
}

func RegisterP2[Actor any, V1 any, V2 any](e framework.ISerialize, f framework.P2Func[Actor, V1, V2]) {
	serviceObj.Register(entity.NewP2Handler(e, f))
}

func RegisterV1[Actor any, V1 any](e framework.ISerialize, f framework.V1Func[Actor, V1]) {
	serviceObj.Register(entity.NewV1Handler(e, f))
}

func RegisterV2[Actor any, V1 any, V2 any](e framework.ISerialize, f framework.V2Func[Actor, V1, V2]) {
	serviceObj.Register(entity.NewV2Handler(e, f))
}
