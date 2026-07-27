package app

import (
	"fmt"
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
	"github.com/hechh/framework/packet"
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

// getTypes 从 Config 提取 nodeType 列表
func getTypes(cfg *config.Config) []uint32 {
	types := make([]uint32, 0, len(cfg.Types))
	for t := range cfg.Types {
		types = append(types, t)
	}
	return types
}

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
}

// NewApp 创建应用实例
func NewApp() *App {
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

// Init 按顺序初始化各组件
func (d *App) Init(cfgPath string, nodeType, nodeId uint32) error {
	// 1. 加载配置（同时初始化 Types/Node/Self 等）
	cfg, err := config.Init(cfgPath, nodeType, nodeId, global.NodeConvertor)
	if err != nil {
		return err
	}
	d.cfg = cfg

	// 2. 日志
	if err := d.logger.Init(cfg.Logger); err != nil {
		return fmt.Errorf("logger init: %w", err)
	}
	mlog.SetObject(d.logger)

	// 3. GC
	if err := d.gc.Init(); err != nil {
		return fmt.Errorf("gc init: %w", err)
	}
	gc.SetObject(d.gc)

	// 4. HTTP 客户端
	if err := d.httpCli.Init(cfg.HttpCli); err != nil {
		return fmt.Errorf("httpcli init: %w", err)
	}
	httpcli.SetObject(d.httpCli)

	// 5. 定时器
	if err := d.timer.Init(cfg.Timer); err != nil {
		return fmt.Errorf("timer init: %w", err)
	}
	timer.SetObject(d.timer)

	// 6. pprof
	if cfg.Node != nil {
		pprofPort := cfg.Node.Port + 10000
		if err := d.pprof.Init(pprofPort); err != nil {
			return fmt.Errorf("pprof init: %w", err)
		}
		pprof.SetObject(d.pprof)
	}

	// 7. 数据库连接池
	if cfg.Mysql != nil {
		if err := d.dbpool.Init(cfg.Mysql); err != nil {
			return fmt.Errorf("dbpool init: %w", err)
		}
		dbpool.SetObject(d.dbpool)
	}

	// 8. Redis 连接池
	if cfg.Redis != nil {
		if err := d.redispool.Init(cfg.Redis); err != nil {
			return fmt.Errorf("redispool init: %w", err)
		}
		redispool.SetObject(d.redispool)
	}

	// 9. 文件监听
	if cfg.Fwatcher != nil {
		if err := d.fwatcher.Init(cfg.Fwatcher); err != nil {
			return fmt.Errorf("fwatcher init: %w", err)
		}
		fwatcher.SetObject(d.fwatcher)
	}

	// 10. 网络（WebSocket）
	if cfg.Node != nil {
		addr := fmt.Sprintf(":%d", cfg.Node.Port)
		if err := d.network.Init(d.timer, addr); err != nil {
			return fmt.Errorf("network init: %w", err)
		}
		network.SetObject(d.network)
	}

	// 11. 路由
	if cfg.Self != nil {
		types := getTypes(cfg)
		if err := d.router.Init(d.timer, cfg.Self, types); err != nil {
			return fmt.Errorf("router init: %w", err)
		}
		router.SetObject(d.router)
	}

	// 12. 集群
	if cfg.Cluster != nil && cfg.Self != nil {
		types := getTypes(cfg)
		if err := d.cluster.Init(cfg.Cluster, cfg.Self, types); err != nil {
			return fmt.Errorf("cluster init: %w", err)
		}
		cluster.SetObject(d.cluster)
	}

	// 13. 消息总线（nats）
	if cfg.MsgBus != nil {
		if err := d.msgbus.Init(cfg.MsgBus, nodeType, nodeId); err != nil {
			return fmt.Errorf("msgbus init: %w", err)
		}
		msgbus.SetObject(d.msgbus)
	}

	mlog.Infof("[APP] 应用初始化完成")
	return nil
}

// Run 启动信号监听，阻塞等待退出信号
func (d *App) Run(sigs ...os.Signal) {
	sigs = append(sigs, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	<-sigChan
	d.Close()
}

// Close 逆序关闭各组件
func (d *App) Close() {
	d.closeComponents()
	mlog.Infof("[APP] 应用已关闭")
}

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

// closeComponents 逆序关闭各组件（与 Init 顺序相反）
func (d *App) closeComponents() {
	if d.msgbus != nil {
		d.msgbus.Close()
	}
	if d.cluster != nil {
		d.cluster.Close()
	}
	if d.router != nil {
		d.router.Close()
	}
	if d.network != nil {
		d.network.Close()
	}
	if d.fwatcher != nil {
		d.fwatcher.Close()
	}
	if d.redispool != nil {
		d.redispool.Close()
	}
	if d.dbpool != nil {
		d.dbpool.Close()
	}
	if d.pprof != nil {
		d.pprof.Close()
	}
	if d.timer != nil {
		d.timer.Close()
	}
	if d.httpCli != nil {
		d.httpCli.Close()
	}
	if d.gc != nil {
		d.gc.Close()
	}
	if d.logger != nil {
		d.logger.Close()
	}
}

// 确保 *packet.Node 实现 cluster.INode 接口
var _ cluster.INode = (*packet.Node)(nil)
