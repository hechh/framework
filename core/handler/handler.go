package handler

import (
	"framework/core"
	"framework/core/handler/internal/handler"
	"framework/core/handler/internal/service"
	"strings"
)

var (
	serviceObj = service.NewService()
)

func init() {
	core.SetGetHandler(Get)
	core.SetGetHandlerByCmd(GetByCmd)
	core.SetGetHandlerByRpc(GetByRpc)
}

func Name2Id(name string) (uint32, bool) {
	return serviceObj.Name2Id(name)
}

func Id2Name(val uint32) (string, bool) {
	return serviceObj.Id2Name(val)
}

func Get(names ...string) core.IHandler {
	return serviceObj.Get(strings.Join(names, "."))
}

func GetByCmd(cmd uint32) core.IHandler {
	return serviceObj.GetByCmd(cmd)
}

func GetByRpc(nodeType uint32, id any) core.IHandler {
	return serviceObj.GetByRpc(nodeType, id)
}

// 注册proto参数请求
func RegisterEvent[Actor any, V1 any](nodeType uint32, cmd uint32, f core.P1Func[Actor, V1]) {
	serviceObj.Register(handler.NewP1Handler(&handler.ProtoEncoder{}, nodeType, cmd, f))
}

func RegisterCmd[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f core.P2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewReqHandler(&handler.ProtoEncoder{}, nodeType, cmd, f))
}

// 注册指针参数请求
func RegisterP1[Actor any, V1 any](nodeType uint32, cmd uint32, f core.P1Func[Actor, V1]) {
	serviceObj.Register(handler.NewP1Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}
func RegisterP2[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f core.P2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewP2Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}
func RegisterP3[Actor any, V1 any, V2 any, V3 any](nodeType uint32, cmd uint32, f core.P3Func[Actor, V1, V2, V3]) {
	serviceObj.Register(handler.NewP3Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}

// 注册基础参数请求
func RegisterV1[Actor any, V1 any](nodeType uint32, cmd uint32, f core.V1Func[Actor, V1]) {
	serviceObj.Register(handler.NewV1Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}
func RegisterV2[Actor any, V1 any, V2 any](nodeType uint32, cmd uint32, f core.V2Func[Actor, V1, V2]) {
	serviceObj.Register(handler.NewV2Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}
func RegisterV3[Actor any, V1 any, V2 any, V3 any](nodeType uint32, cmd uint32, f core.V3Func[Actor, V1, V2, V3]) {
	serviceObj.Register(handler.NewV3Handler(&handler.GobEncoder{}, nodeType, cmd, f))
}
