package fwatcher

import (
	"fmt"

	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *FWatcher
	IsSync bool
	Config *Config
}

func (d *Component) Init() error {
	// 配置缺少 fwatcher 段时 Config 为 nil：此处拦截并给出明确提示，否则下方 IsSync 赋值直接 panic
	if d.Config == nil {
		err := fmt.Errorf("配置为空")
		mlog.Errorf("[fwatcher] 初始化失败，error:%v", err)
		return err
	}
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
