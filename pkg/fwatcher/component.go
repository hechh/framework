package fwatcher

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *FWatcher
	IsSync bool
	Config *Config
}

func (d *Component) Init() error {
	// 初始化模块
	d.Config.IsSync = d.IsSync
	if err := d.Object.Init(d.Config); err != nil {
		mlog.Errorf("[fwatcher] 初始化失败，error:%v", err)
		return err
	}
	SetObject(d.Object)
	mlog.Infof("[fwatcher] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[fwatcher] 关闭成功")
	d.Object = nil
}
