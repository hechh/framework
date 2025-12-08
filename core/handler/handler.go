package handler

import (
	"framework/core/handler/domain"
	"framework/core/handler/internal/base"
	"framework/core/handler/internal/handler"
	"framework/core/handler/internal/service"
)

var (
	serviceObj = service.NewService()
)

func Name2Id(name string) (uint32, bool) {
	return serviceObj.Name2Id(name)
}

func Id2Name(val uint32) (string, bool) {
	return serviceObj.Id2Name(val)
}

func Get(actorFunc string) domain.IHandler {
	return serviceObj.Get(actorFunc)
}

func GetByCmd(cmd uint32) domain.IHandler {
	return serviceObj.GetByCmd(cmd)
}

func GetByRpc(nodeType uint32, id any) domain.IHandler {
	return serviceObj.GetByRpc(nodeType, id)
}

// 注册proto参数请求
func RegisterPb1[Actor any, V1 any](nodeType uint32, cmd uint32, f domain.P1Func[Actor, V1]) {
	serviceObj.Register(handler.NewP1Handler(&base.ProtoEncoder{}, nodeType, cmd, f))
}
func RegisterPb2[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f domain.P2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewP2Handler(&base.ProtoEncoder{}, nodeType, cmd, f))
}

// 注册指针参数请求
func RegisterP1[Actor any, V1 any](nodeType uint32, cmd uint32, f domain.P1Func[Actor, V1]) {
	serviceObj.Register(handler.NewP1Handler(&base.GobEncoder{}, nodeType, cmd, f))
}
func RegisterP2[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f domain.P2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewP2Handler(&base.GobEncoder{}, nodeType, cmd, f))
}
func RegisterP3[Actor any, V1 any, V2 any, V3 any](nodeType uint32, cmd uint32, f domain.P3Func[Actor, V1, V2, V3]) {
	serviceObj.Register(handler.NewP3Handler(&base.GobEncoder{}, nodeType, cmd, f))
}

// 注册基础参数请求
func RegisterV1[Actor any, V1 any](nodeType uint32, cmd uint32, f domain.V1Func[Actor, V1]) {
	serviceObj.Register(handler.NewV1Handler(&base.GobEncoder{}, nodeType, cmd, f))
}
func RegisterV2[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f domain.V2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewV2Handler(&base.GobEncoder{}, nodeType, cmd, f))
}
func RegisterV3[Actor any, V1 any, V2 any, V3 any](nodeType uint32, cmd uint32, f domain.V3Func[Actor, V1, V2, V3]) {
	serviceObj.Register(handler.NewV3Handler(&base.GobEncoder{}, nodeType, cmd, f))
}
