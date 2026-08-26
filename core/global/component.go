package global

import (
	"fmt"

	"github.com/hechh/framework/library/enum"
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Config struct {
	Addr string `yaml:"addr,omitempty"`
	Ip   string `yaml:"ip,omitempty"`   // ip 地址
	Port int32  `yaml:"port,omitempty"` // 端口
}

type Component struct {
	CmdConv  enum.IConvertor
	NodeConv enum.IConvertor
	Gateway  enum.IEnum
	Node     *packet.Node
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	id := fmt.Sprintf("node%d", d.Node.Id)
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, d.Node.Name, id); err != nil {
		mlog.Errorf("[global] 配置加载失败 error:%v", err)
		return err
	}

	// 初始化
	self = d.Node
	self.Addr = cfg.Addr
	self.Ip = cfg.Ip
	self.Port = cfg.Port

	gateway = d.Gateway
	cmdConvertor = d.CmdConv
	nodeConvertor = d.NodeConv
	nodeTypes = d.NodeConv.GetTypes()
	mlog.Infof("[global] 初始化成功")
	return nil
}

func (d *Component) Close() {
	mlog.Infof("[global] 关闭成功")
}
