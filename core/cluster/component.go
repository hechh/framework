package cluster

import (
	"fmt"

	"github.com/hechh/framework/core/global"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Cluster
	Config *Config
}

func (d *Component) Init() error {
	// 配置缺少 discovery 段时 Config 为 nil：此处拦截并给出明确提示，否则服务发现适配器内部解引用 panic
	if d.Config == nil {
		err := fmt.Errorf("配置为空")
		mlog.Errorf("[cluster] 初始化失败，error:%v", err)
		return err
	}
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
