package router

import (
	"github.com/hechh/framework"
	"github.com/hechh/framework/router/internal/service"
	"github.com/hechh/library/yaml"
)

var (
	serviceObj = service.NewService()
)

func init() {
	framework.SetRouter(Get, GetOrNew)
}

func Init(cfg *yaml.NodeConfig, f framework.SaveRouterFunc) {
	serviceObj.Init(cfg, f)
}

func Close() {
	serviceObj.Close()
}

func Get(idType uint32, id uint64) framework.IRouter {
	return serviceObj.Get(idType, id)
}

func GetOrNew(idType uint32, id uint64) framework.IRouter {
	return serviceObj.GetOrNew(idType, id)
}

func Add(str string) error {
	return serviceObj.Add(str)
}
