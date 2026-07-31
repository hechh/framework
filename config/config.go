package config

import (
	"strings"

	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/enum"
	"github.com/hechh/library/base/fileutil"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/timer"
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
	types    map[uint32]map[uint32]*NodeConfig `yaml:"-"`
	node     *NodeConfig                       `yaml:"-"`
	gateway  *NodeConfig                       `yaml:"-"`
	self     *packet.Node                      `yaml:"-"`
}

func (d *Config) GetSupports() []uint32 {
	return templ.Map2List(d.types)
}

func (d *Config) GetGateway() *NodeConfig {
	return d.gateway
}

func (d *Config) GetSelfNode() *packet.Node {
	return d.self
}

func (d *Config) GetSelfConfig() *NodeConfig {
	return d.node
}

func (d *Config) GetNodeConfig(nodeType, nodeId uint32) *NodeConfig {
	if items, ok := d.types[nodeType]; ok {
		return items[nodeId]
	}
	return nil
}

func (d *Config) Init(filename string, nodeType, nodeId uint32, conv enum.IConvertor) error {
	// 加载配置
	if err := fileutil.LoadYaml(filename, d); err != nil {
		return err
	}

	d.Mysql.UidModValue = d.Common.UidModValue
	d.Redis.UidModValue = d.Common.UidModValue

	d.types = make(map[uint32]map[uint32]*NodeConfig)
	for nodeName, items := range d.Nodes {
		for nodeId, cfg := range items {
			cfg.Type = conv.ToUint32(strings.ToUpper(nodeName))
			cfg.Id = nodeId
			cfg.Name = nodeName

			// 支持的节点类型
			if _, ok := d.types[cfg.Type]; !ok {
				d.types[cfg.Type] = make(map[uint32]*NodeConfig)
			}
			d.types[cfg.Type][cfg.Id] = cfg

			// 是否为网关节点
			if cfg.Type&define.GATEWAY_MASK == define.GATEWAY_MASK {
				d.gateway = cfg
			}

			// 当前节点
			if cfg.Type == nodeType && cfg.Id == nodeId {
				d.node = cfg
				d.self = &packet.Node{
					Type: cfg.Type,
					Id:   cfg.Id,
					Name: cfg.Name,
					Ip:   cfg.Ip,
					Port: int32(cfg.Port),
				}
			}

			// 节点日志
			if d.Logger != nil && cfg.Logger == nil {
				cfg.Logger = d.Logger
			}
			if cfg.Logger != nil && d.Logger != nil {
				cfg.Logger = &mlog.Config{
					Mode:      templ.Or(cfg.Logger.Mode == "", d.Logger.Mode, cfg.Logger.Mode),
					Path:      templ.Or(cfg.Logger.Path == "", d.Logger.Path, cfg.Logger.Path),
					Level:     templ.Or(cfg.Logger.Level == "", d.Logger.Level, cfg.Logger.Level),
					Format:    templ.Or(cfg.Logger.Format == "", d.Logger.Format, cfg.Logger.Format),
					Name:      cfg.Name,
					IsCaller:  templ.Or(cfg.Logger.IsCaller, cfg.Logger.IsCaller, d.Logger.IsCaller),
					CacheSize: templ.Or(cfg.Logger.CacheSize == 0, d.Logger.CacheSize, cfg.Logger.CacheSize),
				}
			}
		}
	}
	return nil
}
