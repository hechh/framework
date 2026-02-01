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

func RegisterRpc[T any, U any](e framework.ISerialize, nodeType, cmd framework.IEnum, name string) {
	serviceObj.RegisterRpc(entity.NewRpc[T, U](e, nodeType.Integer(), cmd.Integer(), name))
}

func Register0[Actor any](e framework.ISerialize, f framework.EmptyFunc[Actor]) {
	serviceObj.Register(entity.NewEmptyHandler(e, f))
}

func RegisterP1[Actor any, N any](e framework.ISerialize, f framework.P1Func[Actor, N]) {
	serviceObj.Register(entity.NewP1Handler(e, f))
}

func RegisterP2[Actor any, N any, R any](e framework.ISerialize, f framework.P2Func[Actor, N, R]) {
	serviceObj.Register(entity.NewP2Handler(e, f))
}

func RegisterV1[Actor any, N any](e framework.ISerialize, f framework.V1Func[Actor, N]) {
	serviceObj.Register(entity.NewV1Handler(e, f))
}

func RegisterV2[Actor any, N any, R any](e framework.ISerialize, f framework.V2Func[Actor, N, R]) {
	serviceObj.Register(entity.NewV2Handler(e, f))
}
