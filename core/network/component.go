package network

import (
	"fmt"

	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/core/network/internal/frame"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *Network
	Node    *packet.Node
	Decoder func([]byte) (*packet.Packet, error)
	Encoder func(*packet.Packet) ([]byte, error)
	Handler func(*packet.Packet) error
}

func (d *Component) Init(data map[string]any) error {
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

	self := global.GetSelf()
	addr := fmt.Sprintf(":%d", self.Port)

	// 初始化模块
	if err := d.Object.Init(addr); err != nil {
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
