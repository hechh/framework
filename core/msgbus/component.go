package msgbus

import (
	"fmt"

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
	// 配置缺少 msgbus 段时 Config 为 nil：此处拦截并给出明确提示，否则消息总线适配器内部解引用 panic
	if d.Config == nil {
		err := fmt.Errorf("配置为空")
		mlog.Errorf("[msgbus] 初始化失败，error:%v", err)
		return err
	}
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
