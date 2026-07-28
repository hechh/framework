package mockapp

import (
	"github.com/hechh/framework/app"
	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/cluster/adapter/mockdis"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/msgbus/adapter/mockbus"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/core/network/adapter/mocknet"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/dbpool/adapter/mockdb"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/fwatcher/adapter/mocksync"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/redispool/adapter/mockredis"
)

type App struct {
	*app.App
}

func New() *App {
	// 设置mock模式
	mockApp := app.New()
	mockApp.SetDbPool(dbpool.NewDbPool(mockdb.New))
	mockApp.SetFwatcher(fwatcher.NewFwatcher(mocksync.NewMonitor))
	mockApp.SetRedisPool(redispool.NewRedisPool(mockredis.New))
	mockApp.SetNetwork(network.NewNetwork(mocknet.New))
	mockApp.SetMsgBus(msgbus.NewMsgBus(mockbus.New()))
	mockApp.SetCluster(cluster.NewCluster(mockdis.New()))
	// 注册全部组件
	mockApp.Register(&app.Logger{})
	mockApp.Register(&app.HttpCli{})
	mockApp.Register(&app.Gc{})
	mockApp.Register(&app.Pprof{})
	mockApp.Register(&app.Timer{})
	mockApp.Register(&app.DbPool{})
	mockApp.Register(&app.RedisPool{})
	mockApp.Register(&app.Fwatcher{})
	mockApp.Register(&app.Router{})
	mockApp.Register(&app.Cluster{})
	mockApp.Register(&app.MsgBus{})
	mockApp.Register(&app.Network{})
	return &App{App: mockApp}
}
