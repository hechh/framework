package network

import (
	"fmt"

	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/network/internal/frame"
)

type Component struct {
	Object  *Network
	Node    *packet.Node
	Decoder func([]byte) (*packet.Packet, error)
	Encoder func(*packet.Packet) ([]byte, error)
	Handler func(*packet.Packet) error
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	id := fmt.Sprintf("node%d", d.Node.Id)
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, d.Node.Name, id); err != nil {
		mlog.Errorf("[network] 配置加载失败 error:%v", err)
		return err
	}

	// 设置处理器
	if d.Decoder == nil {
		d.Decoder = frame.Decode
	}
	if d.Encoder == nil {
		d.Encoder = frame.Encode
	}
	SetDecodeFunc(d.Decoder)
	SetEncodeFunc(d.Encoder)
	SetPacketFunc(d.Handler)

	// 初始化模块
	if err := d.Object.Init(cfg); err != nil {
		mlog.Errorf("[network] 初始化失败，error:%v", err)
		return err
	}
	SetObject(d.Object)
	mlog.Infof("[network] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[network] 关闭成功")
	d.Object = nil
}
