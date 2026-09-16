package msgbus

import (
	"github.com/hechh/framework/core/router"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *MsgBus
	Handler func(*packet.Packet)
	Config  *Config
}

func (d *Component) Init() error {
	SetPacketFunc(router.RouteHandler(d.Handler))

	if err := d.Object.Init(d.Config); err != nil {
		mlog.Errorf("[msgbus] 初始化失败，error:%v", err)
		return err
	}

	SetObject(d.Object)
	mlog.Infof("[msgbus] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[msgbus] 关闭成功")
	d.Object = nil
}
