package config

import (
	"os"
	"strings"

	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/define"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/base/utils"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/timer"
	"go.yaml.in/yaml/v2"
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

func Load(filename string, nodeType uint32, nodeId uint32, conv utils.IConvertor) (*Config, error) {
	// 加载配置
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	gcfg := new(Config)
	if err := yaml.Unmarshal(content, gcfg); err != nil {
		return nil, err
	}

	gcfg.Mysql.UidModValue = gcfg.Common.UidModValue
	gcfg.Redis.UidModValue = gcfg.Common.UidModValue

	gcfg.Types = make(map[uint32]map[uint32]*NodeConfig)
	for nodeName, items := range gcfg.Nodes {
		for nodeId, cfg := range items {
			cfg.Type = conv.ToUint32(strings.ToUpper(nodeName))
			cfg.Id = nodeId
			cfg.Name = nodeName

			// 支持的节点类型
			if _, ok := gcfg.Types[cfg.Type]; !ok {
				gcfg.Types[cfg.Type] = make(map[uint32]*NodeConfig)
			}
			gcfg.Types[cfg.Type][cfg.Id] = cfg

			// 是否为网关节点
			if cfg.Type&define.GATEWAY_MASK == define.GATEWAY_MASK {
				gcfg.Gateway = cfg
			}

			// 当前节点
			if cfg.Type == nodeType && cfg.Id == nodeId {
				gcfg.Node = cfg
				gcfg.Self = &packet.Node{
					Type: cfg.Type,
					Id:   cfg.Id,
					Name: cfg.Name,
					Ip:   cfg.Ip,
					Port: int32(cfg.Port),
				}
			}

			// 节点日志
			if gcfg.Logger != nil && cfg.Logger == nil {
				cfg.Logger = gcfg.Logger
			}
			if cfg.Logger != nil && gcfg.Logger != nil {
				cfg.Logger = &mlog.Config{
					Mode:     templ.Or(cfg.Logger.Mode == "", gcfg.Logger.Mode, cfg.Logger.Mode),
					Path:     templ.Or(cfg.Logger.Path == "", gcfg.Logger.Path, cfg.Logger.Path),
					Level:    templ.Or(cfg.Logger.Level == "", gcfg.Logger.Level, cfg.Logger.Level),
					Format:   templ.Or(cfg.Logger.Format == "", gcfg.Logger.Format, cfg.Logger.Format),
					Name:     cfg.Name,
					IsCaller: templ.Or(cfg.Logger.IsCaller, cfg.Logger.IsCaller, gcfg.Logger.IsCaller),
					Cache:    templ.Or(cfg.Logger.Cache == 0, gcfg.Logger.Cache, cfg.Logger.Cache),
				}
			}
		}
	}
	return gcfg, nil
}
