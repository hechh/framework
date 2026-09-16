package cluster

import (
	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Cluster
	Config *Config
}

func (d *Component) Init() error {
	if err := d.Object.Init(d.Config, global.GetSelf(), global.GetSupportNodeTypes()); err != nil {
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
