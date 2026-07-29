package service

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
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/pprof"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/redispool/adapter/goredis"
	"github.com/hechh/library/timer"
	"github.com/hechh/library/timer/adapter/lockfree_timer"
)

type IService interface {
	Init(...msgqueue.Option) error
	Close()
}

type IComponent interface {
	Init(*Service) error
	Close(*Service)
}

// Service 应用生命周期管理器
type Service struct {
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
	services  []IService
}

// NewApp 创建应用实例
func New() *Service {
	return &Service{
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

// 注册组件
func (d *Service) Register(c any) {
	switch vv := c.(type) {
	case IComponent:
		d.services = append(d.services, &Wrapper{app: d, comp: vv})
	case IService:
		d.services = append(d.services, vv)
	}
}

// ==================== Setter（DI 注入，可覆盖默认组件） ====================
func (d *Service) SetConfig(cfg *config.Config)        { d.cfg = cfg }
func (d *Service) SetDbPool(v *dbpool.DbPool)          { d.dbpool = v }
func (d *Service) SetFwatcher(v *fwatcher.Fwatcher)    { d.fwatcher = v }
func (d *Service) SetGc(v *gc.Gc)                      { d.gc = v }
func (d *Service) SetHttpCli(v *httpcli.HttpClient)    { d.httpCli = v }
func (d *Service) SetLogger(v *mlog.Logger)            { d.logger = v }
func (d *Service) SetPprof(v *pprof.Pprof)             { d.pprof = v }
func (d *Service) SetRedisPool(v *redispool.RedisPool) { d.redispool = v }
func (d *Service) SetTimer(v *timer.Timer)             { d.timer = v }
func (d *Service) SetRouter(v *router.Router)          { d.router = v }
func (d *Service) SetNetwork(v *network.Network)       { d.network = v }
func (d *Service) SetMsgBus(v *msgbus.MsgBus)          { d.msgbus = v }
func (d *Service) SetCluster(v *cluster.Cluster)       { d.cluster = v }

// ==================== Getter ====================
func (d *Service) GetConfig() *config.Config          { return d.cfg }
func (d *Service) GetDbPool() *dbpool.DbPool          { return d.dbpool }
func (d *Service) GetFwatcher() *fwatcher.Fwatcher    { return d.fwatcher }
func (d *Service) GetGc() *gc.Gc                      { return d.gc }
func (d *Service) GetHttpCli() *httpcli.HttpClient    { return d.httpCli }
func (d *Service) GetLogger() *mlog.Logger            { return d.logger }
func (d *Service) GetPprof() *pprof.Pprof             { return d.pprof }
func (d *Service) GetRedisPool() *redispool.RedisPool { return d.redispool }
func (d *Service) GetTimer() *timer.Timer             { return d.timer }
func (d *Service) GetRouter() *router.Router          { return d.router }
func (d *Service) GetNetwork() *network.Network       { return d.network }
func (d *Service) GetMsgBus() *msgbus.MsgBus          { return d.msgbus }
func (d *Service) GetCluster() *cluster.Cluster       { return d.cluster }

// Init 按顺序初始化各组件
func (d *Service) Init(filename string, nodeType, nodeId uint32) error {
	// 1. 加载配置（同时初始化 Types/Node/Self 等）
	cfg, err := config.Load(filename, nodeType, nodeId, global.NodeConvertor)
	if err != nil {
		return err
	}
	d.cfg = cfg

	global.Self = cfg.Self
	global.GatewayNodeType = cfg.Gateway.Type

	for i, comp := range d.services {
		if err := comp.Init(); err != nil {
			for j := i - 1; j >= 0; j-- {
				d.services[j].Close()
			}
			return err
		}
	}

	return nil
}

// Close 逆序关闭各组件
func (d *Service) Close() {
	for j := len(d.services) - 1; j >= 0; j-- {
		d.services[j].Close()
	}
}

// Run 启动信号监听，阻塞等待退出信号
func (d *Service) Run(sigs ...os.Signal) {
	sigs = append(sigs, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	<-sigChan
	d.Close()
}
