package handler

import (
	"framework/define"
	"framework/internal/handler"
	"framework/repository/handler/domain"
	"framework/repository/handler/internal/entity"
	"framework/repository/handler/internal/manager"
)

func init() {
	handler.SetHandler(manager.Name2Id, manager.Id2Name, manager.GetRpc, manager.GetCmd, manager.Get)
}

func RegisterL1[A any, V any](nodeType int32, cmd int32, method domain.L1Func[A, V]) {
	manager.Register(entity.NewL1Handler(nodeType, cmd, method))
}

func RegisterL2[A any, V1 any, V2 any](nodeType int32, cmd int32, method domain.L2Func[A, V1, V2]) {
	manager.Register(entity.NewL2Handler(nodeType, cmd, method))
}

// 注册define.IHandler
func RegisterZ0[A any](nodeType int32, cmd int32, method domain.Z0Func[A]) {
	manager.Register(entity.NewZ0Handler(nodeType, cmd, method))
}

func RegisterZ1[A any](nodeType int32, cmd int32, method domain.Z1Func[A]) {
	manager.Register(entity.NewZ1Handler(nodeType, cmd, method))
}

func RegisterP1[A any, V any](nodeType int32, cmd int32, method domain.P1Func[A, V]) {
	manager.Register(entity.NewP1Handler(nodeType, cmd, method))
}

func RegisterP2[A any, V1 any, V2 any](nodeType int32, cmd int32, method domain.P2Func[A, V1, V2]) {
	manager.Register(entity.NewP2Handler(nodeType, cmd, method))
}

func RegisterG1[A any, V any](nodeType int32, cmd int32, method domain.G1Func[A, V]) {
	manager.Register(entity.NewG1Handler(nodeType, cmd, method))
}

func RegisterG2[A any, V1 any, V2 any](nodeType int32, cmd int32, method domain.G2Func[A, V1, V2]) {
	manager.Register(entity.NewG2Handler(nodeType, cmd, method))
}

// 注册全局RPC
func RegisterRpcGob(nodeType int32, cmd int32, actorFunc string) {
	manager.RegisterRpc(entity.NewGobRpc(nodeType, cmd, actorFunc))
}

func RegisterRpcProto(nodeType int32, cmd int32, actorFunc string) {
	manager.RegisterRpc(entity.NewProtoRpc(nodeType, cmd, actorFunc))
}

func GetByCmd(cmd int32) define.IHandler {
	return manager.GetCmd(cmd)
}
