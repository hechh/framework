package router

import (
	"framework/define"
	"framework/library/yaml"
	"framework/repository/router/domain"
	"framework/repository/router/internal/entity"
	"framework/repository/router/internal/service"
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

func Get(idType uint32, id uint64) define.IRouter {
	return serviceObj.Get(idType, id)
}

func GetOrNew(idType uint32, id uint64) define.IRouter {
	return serviceObj.GetOrNew(idType, id)
}
