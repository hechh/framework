package msgbus

import (
	"github.com/hechh/framework/core/router"
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *MsgBus
	Handler func(*packet.Packet)
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "msgbus"); err != nil {
		mlog.Errorf("[msgbus] 配置加载失败 error:%v", err)
		return err
	}

	SetPacketFunc(router.RouteHandler(d.Handler))

	if err := d.Object.Init(cfg); err != nil {
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
