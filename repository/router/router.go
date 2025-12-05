package router

import (
	"framework/define"
	"framework/internal/router"
	"framework/library/yaml"
	"framework/repository/router/internal/entity"
	"framework/repository/router/internal/service"
)

var (
	routerObj = service.NewRouterService(entity.NewRouter, filter)
)

func filter(r define.IRouter) bool {
	return r.GetType() != 0
}

func Init(cfg *yaml.NodeConfig, client define.IRedis) {
	routerObj.Init(cfg, client)
}

func Close() {
	routerObj.Close()
}

func init() {
	router.Set(routerObj.Load, routerObj.LoadOrNew)
}
