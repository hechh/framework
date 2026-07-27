package mock

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

// Init 使用 mock/stub 实现替换 App 中的组件，避免依赖外部服务（MySQL、Redis、etcd、NATS 等）。
func Init(app *app.App) {
	// 数据库连接池 → SQLite（内存模式，无需 MySQL）
	app.SetDbPool(dbpool.NewDbPool(mockdb.New))

	// 文件监听 → 嵌入式 etcd（进程内，无需外部 etcd）
	app.SetFwatcher(fwatcher.NewFwatcher(mocksync.NewMonitor))

	// Redis 连接池 → miniredis（进程内，无需外部 Redis）
	app.SetRedisPool(redispool.NewRedisPool(mockredis.New))

	// 网络层 → httptest + WebSocket mock（进程内，无需真实端口）
	app.SetNetwork(network.NewNetwork(mocknet.New))

	// 消息总线 → 嵌入式 NATS（进程内，无需外部 NATS）
	app.SetMsgBus(msgbus.NewMsgBus(mockbus.New()))

	// 集群发现 → 嵌入式 etcd（进程内，无需外部 etcd）
	app.SetCluster(cluster.NewCluster(mockdis.New()))
}
