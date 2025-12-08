package router

import (
	"framework/core/router/domain"
	"framework/core/router/internal/entity"
	"framework/core/router/internal/service"
	"framework/library/yaml"
)

var (
	serviceObj = service.NewService(entity.NewRouter)
)

func Init(cfg *yaml.NodeConfig, ff domain.FilterFunc) {
	serviceObj.Init(cfg, ff)
}

func Close() {
	serviceObj.Close()
}

func Get(idType uint32, id uint64) domain.IRouter {
	return serviceObj.Get(idType, id)
}

func GetOrNew(idType uint32, id uint64) domain.IRouter {
	return serviceObj.GetOrNew(idType, id)
}
