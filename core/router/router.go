package router

import (
	"framework/core"
	"framework/core/router/internal/entity"
	"framework/core/router/internal/service"
	"framework/library/yaml"
)

var (
	serviceObj = service.NewService(entity.NewRouter)
)

func init() {
	core.SetGetRouter(Get)
	core.SetGetOrNewRouter(GetOrNew)
}

func Init(cfg *yaml.NodeConfig) {
	serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func Get(idType uint32, id uint64) core.IRouter {
	return serviceObj.Get(idType, id)
}

func GetOrNew(idType uint32, id uint64) core.IRouter {
	return serviceObj.GetOrNew(idType, id)
}
