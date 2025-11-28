package router

import (
	"framework/define"
	"framework/internal/router"
	"framework/library/yaml"
	"framework/repository/router/internal/entity"
	"framework/repository/router/internal/service"
)

var (
	routerObj = service.NewRouterService(entity.NewRouter)
)

func Init(cfg *yaml.NodeConfig, client define.IRedis, idType int32) {
	routerObj.Init(cfg, client, idType)
}

func Close() {
	routerObj.Close()
}

func init() {
	router.Set(routerObj.Load, routerObj.LoadOrNew)
}
