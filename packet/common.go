package packet

import (
	"sort"
	sync "sync"
)

type Config struct {
	Redis     *RedisConfig                      `yaml:"redis"`
	Mysql     *MysqlConfig                      `yaml:"mysql"`
	MsgQueue  *MsgQueueConfig                   `yaml:"msgqueue"`
	Discovery *DiscoveryConfig                  `yaml:"discovery"`
	Common    *CommonConfig                     `yaml:"common"`
	Logger    *LoggerConfig                     `yaml:"logger"`
	Table     *TableConfig                      `yaml:"table"`
	Nodes     map[string]map[uint32]*NodeConfig `yaml:"nodes"`
	Types     map[uint32]map[uint32]*NodeConfig `yaml:"-"`
	Node      *NodeConfig                       `yaml:"-"`
	Gateway   *NodeConfig                       `yaml:"-"`
}

func (c *Config) GetNodeConfigs(nodeType uint32) (rets []*NodeConfig) {
	for _, cfg := range c.Types[nodeType] {
		rets = append(rets, cfg)
	}
	sort.Slice(rets, func(i, j int) bool { return rets[i].Id < rets[j].Id })
	return
}

func (c *Config) RandGatewayConfig(uid uint64) *NodeConfig {
	rets := c.GetNodeConfigs(c.Gateway.Type)
	ll := len(rets)
	if ll <= 0 {
		return nil
	}
	return rets[uid%uint64(ll)]
}

var (
	headPool = sync.Pool{
		New: func() any { return new(Head) },
	}
)

func GetHead(opts ...func(*Head)) *Head {
	obj := headPool.Get().(*Head)
	for _, opt := range opts {
		opt(obj)
	}
	return obj
}

func PutHead(val *Head) {
	*val = Head{}
	headPool.Put(val)
}
