package httpcli

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	object *HttpClient
	Config *Config
}

func (d *Component) Init() error {
	// 初始化模块
	d.object = NewHttpClient()
	if err := d.object.Init(d.Config); err != nil {
		mlog.Errorf("[httpcli] 初始化失败，error:%v", err)
		return err
	}

	SetObject(d.object)
	mlog.Infof("[httpcli] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.object != nil {
		d.object.Close()
	}
	mlog.Infof("[httpcli] 关闭成功")
	d.object = nil
}
