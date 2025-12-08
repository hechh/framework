package router

import (
	"framework/define"
	"framework/internal/router"
	"framework/library/yaml"
	"framework/repository/router/internal/entity"
	"framework/repository/router/internal/service"
)

var (
	obj *service.RouterService
)

func filter(r define.IRouter) bool {
	return r.GetType() != 0
}

func Init(cfg *yaml.NodeConfig, client define.IRedis) {
	obj = service.NewRouterService(entity.NewRouter, filter)
	obj.Init(cfg, client)
	router.Set(obj.Load, obj.LoadOrNew)
}

func Close() {
	obj.Close()
}
