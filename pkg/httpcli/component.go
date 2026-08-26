package httpcli

import (
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	object *HttpClient
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "httpcli"); err != nil {
		mlog.Errorf("[httpcli] 配置加载失败 error:%v", err)
		return err
	}

	// 初始化模块
	d.object = NewHttpClient()
	if err := d.object.Init(cfg); err != nil {
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
