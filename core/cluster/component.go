package cluster

import (
	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Cluster
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "discovery"); err != nil {
		mlog.Errorf("[cluster] 配置加载失败 error:%v", err)
		return err
	}

	if err := d.Object.Init(cfg, global.GetSelf(), global.GetSupportNodeTypes()); err != nil {
		mlog.Errorf("[cluster] 初始化失败，error:%v", err)
		return err
	}
	SetObject(d.Object)
	mlog.Infof("[cluster] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[cluster] 关闭成功")
	d.Object = nil
}
