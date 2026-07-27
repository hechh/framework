package global

import (
	"sync"

	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/utils"
)

var (
	CmdConvertor    utils.IConvertor
	NodeConvertor   utils.IConvertor
	Self            *packet.Node
	GatewayNodeType uint32
	headPool        = sync.Pool{
		New: func() any { return new(packet.Head) },
	}
)

func SetNodeConvertor(n map[string]int32, i map[int32]string) {
	NodeConvertor = utils.WrapConvertor(n, i)
}

func SetCmdConvertor(n map[string]int32, i map[int32]string) {
	CmdConvertor = utils.WrapConvertor(n, i)
}

func GetHead(opts ...func(*packet.Head)) *packet.Head {
	obj := headPool.Get().(*packet.Head)
	for _, opt := range opts {
		opt(obj)
	}
	return obj
}

func PutHead(val *packet.Head) {
	*val = packet.Head{}
	headPool.Put(val)
}

/*
func Init(filename string, nodeType uint32, nodeId uint32) (*packet.Config, error) {
	// 加载配置
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	GlobalCfg = new(packet.Config)
	if err := yaml.Unmarshal(content, GlobalCfg); err != nil {
		return nil, err
	}

	GlobalCfg.Types = make(map[uint32]map[uint32]*packet.NodeConfig)
	for nodeName, items := range GlobalCfg.Nodes {
		for nodeId, cfg := range items {
			cfg.Type = NodeConvertor.ToUint32(strings.ToUpper(nodeName))
			cfg.Id = nodeId
			cfg.Name = nodeName
			// 支持的节点类型
			if _, ok := GlobalCfg.Types[cfg.Type]; !ok {
				GlobalCfg.Types[cfg.Type] = make(map[uint32]*packet.NodeConfig)
			}
			GlobalCfg.Types[cfg.Type][cfg.Id] = cfg
			// 是否为网关节点
			if cfg.Type&constant.GATEWAY_MASK == constant.GATEWAY_MASK {
				GlobalCfg.Gateway = cfg
			}
			// 当前节点
			if cfg.Type == nodeType && cfg.Id == nodeId {
				GlobalCfg.Node = cfg
				Self = &packet.Node{
					Type: cfg.Type,
					Id:   cfg.Id,
					Name: cfg.Name,
					Ip:   cfg.Ip,
					Port: cfg.Port,
				}
			}
		}
	}
	return GlobalCfg, nil
}
*/
