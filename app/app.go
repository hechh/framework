package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

// 节点配置
type NodeConfig struct {
	Type   uint32       `yaml:"type,omitempty"`   // 节点类型
	Id     uint32       `yaml:"id,omitempty"`     // 节点id
	Name   string       `yaml:"name,omitempty"`   // 节点名字
	Ip     string       `yaml:"ip,omitempty"`     // ip 地址
	Port   int          `yaml:"port,omitempty"`   // 端口
	Logger *mlog.Config `yaml:"logger,omitempty"` // 日志
}

// 公共配置
type CommonConfig struct {
	Env         string `yaml:"env,omitempty"`
	IsOpenPprof bool   `yaml:"is_open_pprof,omitempty"`
	JwtSecret   string `yaml:"jwt_secret,omitempty"`
	AesSecret   string `yaml:"aes_secret,omitempty"`
	UidModValue uint64 `yaml:"uid_mod_value,omitempty"`
	PayPort     int32  `yaml:"pay_port,omitempty"`
	GeoDbPath   string `yaml:"geo_db_path,omitempty"` // GeoLite2-City.mmdb 文件路径
}

type Config struct {
	Mysql    *dbpool.Config                    `yaml:"mysql,omitempty"`
	Redis    *redispool.Config                 `yaml:"redis,omitempty"`
	Fwatcher *fwatcher.Config                  `yaml:"fwatcher,omitempty"`
	HttpCli  *httpcli.Config                   `yaml:"http_cli,omitempty"`
	Logger   *mlog.Config                      `yaml:"logger,omitempty"`
	Timer    *timer.Config                     `yaml:"timer,omitempty"`
	MsgBus   *msgbus.Config                    `yaml:"msgbus,omitempty"`
	Cluster  *cluster.Config                   `yaml:"cluster,omitempty"`
	Common   *CommonConfig                     `yaml:"common,omitempty"`
	Nodes    map[string]map[uint32]*NodeConfig `yaml:"nodes,omitempty"`
	Types    map[uint32]map[uint32]*NodeConfig `yaml:"-"`
	Node     *NodeConfig                       `yaml:"-"`
	Gateway  *NodeConfig                       `yaml:"-"`
	Self     *packet.Node                      `yaml:"-"`
}

// App 应用生命周期管理器
type App struct {
	cfg       *Config
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

// ==================== Setter（DI 注入，可覆盖默认组件） ====================
func (d *App) SetConfig(cfg *Config)               { d.cfg = cfg }
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
func (d *App) GetConfig() *Config                 { return d.cfg }
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

// Init 按顺序初始化各组件
func (d *App) Init(cfgPath string, nodeType, nodeId uint32) error {
	// 1. 加载配置（同时初始化 Types/Node/Self 等）
	cfg, err := Init(cfgPath, nodeType, nodeId, global.NodeConvertor)
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
