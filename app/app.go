package app

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/hechh/framework/config"
	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/cluster/adapter/discovery"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/msgbus/adapter/nats"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/core/network/adapter/websocket"
	"github.com/hechh/framework/core/router"
	"github.com/hechh/framework/global"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/dbpool/adapter/mysql"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/fwatcher/adapter/etcdsync"
	"github.com/hechh/library/gc"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/pprof"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/redispool/adapter/goredis"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/timer/adapter/lockfree_timer"
)

// App 应用生命周期管理器
type App struct {
	cfg       *config.Config
	dbpool    *dbpool.DbPool
	fwatcher  *fwatcher.Fwatcher
	gc        *gc.Gc
	httpCli   *httpcli.HttpClient
	logger    *mlog.Logger
	pprof     *pprof.Pprof
	redispool *redispool.RedisPool
	timer     *timer.Timer
	router    *router.Router
	network   *network.Network
	msgbus    *msgbus.MsgBus
	cluster   *cluster.Cluster
	comps     []IComponent
}

// NewApp 创建应用实例
func New() *App {
	return &App{
		dbpool:    dbpool.NewDbPool(mysql.NewClient),
		fwatcher:  fwatcher.NewFwatcher(etcdsync.NewEtcdSync),
		gc:        &gc.Gc{},
		httpCli:   httpcli.NewHttpClient(),
		logger:    mlog.NewLogger(),
		pprof:     &pprof.Pprof{},
		redispool: redispool.NewRedisPool(func() *goredis.Client { return &goredis.Client{} }),
		timer:     timer.NewTimer(lockfree_timer.NewTimer),
		router:    router.NewRouter(),
		network:   network.NewNetwork(websocket.NewServer),
		msgbus:    msgbus.NewMsgBus(nats.NewNats()),
		cluster:   cluster.NewCluster(discovery.NewEtcd()),
	}
}

// ==================== Setter（DI 注入，可覆盖默认组件） ====================
func (d *App) SetConfig(cfg *config.Config)        { d.cfg = cfg }
func (d *App) SetDbPool(v *dbpool.DbPool)          { d.dbpool = v }
func (d *App) SetFwatcher(v *fwatcher.Fwatcher)    { d.fwatcher = v }
func (d *App) SetGc(v *gc.Gc)                      { d.gc = v }
func (d *App) SetHttpCli(v *httpcli.HttpClient)    { d.httpCli = v }
func (d *App) SetLogger(v *mlog.Logger)            { d.logger = v }
func (d *App) SetPprof(v *pprof.Pprof)             { d.pprof = v }
func (d *App) SetRedisPool(v *redispool.RedisPool) { d.redispool = v }
func (d *App) SetTimer(v *timer.Timer)             { d.timer = v }
func (d *App) SetRouter(v *router.Router)          { d.router = v }
func (d *App) SetNetwork(v *network.Network)       { d.network = v }
func (d *App) SetMsgBus(v *msgbus.MsgBus)          { d.msgbus = v }
func (d *App) SetCluster(v *cluster.Cluster)       { d.cluster = v }

// ==================== Getter ====================
func (d *App) GetConfig() *config.Config          { return d.cfg }
func (d *App) GetDbPool() *dbpool.DbPool          { return d.dbpool }
func (d *App) GetFwatcher() *fwatcher.Fwatcher    { return d.fwatcher }
func (d *App) GetGc() *gc.Gc                      { return d.gc }
func (d *App) GetHttpCli() *httpcli.HttpClient    { return d.httpCli }
func (d *App) GetLogger() *mlog.Logger            { return d.logger }
func (d *App) GetPprof() *pprof.Pprof             { return d.pprof }
func (d *App) GetRedisPool() *redispool.RedisPool { return d.redispool }
func (d *App) GetTimer() *timer.Timer             { return d.timer }
func (d *App) GetRouter() *router.Router          { return d.router }
func (d *App) GetNetwork() *network.Network       { return d.network }
func (d *App) GetMsgBus() *msgbus.MsgBus          { return d.msgbus }
func (d *App) GetCluster() *cluster.Cluster       { return d.cluster }

// 注册组件
func (d *App) Register(c IComponent) {
	d.comps = append(d.comps, c)
}

// Init 按顺序初始化各组件
func (d *App) Init(filename string, nodeType, nodeId uint32) error {
	// 1. 加载配置（同时初始化 Types/Node/Self 等）
	cfg, err := config.Load(filename, nodeType, nodeId, global.NodeConvertor)
	if err != nil {
		return err
	}
	d.cfg = cfg

	global.Self = cfg.Self
	global.GatewayNodeType = cfg.Gateway.Type

	for i, comp := range d.comps {
		if err := comp.Init(d); err != nil {
			for j := i - 1; j >= 0; j-- {
				d.comps[j].Close(d)
			}
			return err
		}
	}
	return nil
}

// Close 逆序关闭各组件
func (d *App) Close() {
	for j := len(d.comps) - 1; j >= 0; j-- {
		d.comps[j].Close(d)
	}
}

// Run 启动信号监听，阻塞等待退出信号
func (d *App) Run(sigs ...os.Signal) {
	sigs = append(sigs, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	<-sigChan
	d.Close()
}
