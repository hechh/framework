package router

import (
	"framework/core/define"
	"framework/core/router/internal/entity"
	"framework/core/router/internal/service"
	"framework/library/yaml"
)

var (
	serviceObj = service.NewService(entity.NewRouter)
)

func Init(cfg *yaml.NodeConfig) {
	serviceObj.Init(cfg)
}

func Close() {
	serviceObj.Close()
}

func Get(idType uint32, id uint64) define.IRouter {
	return serviceObj.Get(idType, id)
}

func GetOrNew(idType uint32, id uint64) define.IRouter {
	return serviceObj.GetOrNew(idType, id)
}
