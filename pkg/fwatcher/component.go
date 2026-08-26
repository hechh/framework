package fwatcher

import (
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *FWatcher
	IsSync bool
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "fwatcher"); err != nil {
		mlog.Errorf("[fwatcher] 配置加载失败 error:%v", err)
		return err
	}
	cfg.IsSync = d.IsSync

	// 初始化模块
	if err := d.Object.Init(cfg); err != nil {
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
