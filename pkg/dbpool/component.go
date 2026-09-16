package dbpool

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *DbPool
	Configs map[string]*Config
}

// 初始化
func (d *Component) Init() error {
	// 模块初始化
	if err := d.Object.Init(d.Configs); err != nil {
		mlog.Errorf("[dbpool] 初始化失败，error=%v", err)
		return err
	}
	mlog.Infof("[dbpool] 初始化成功")
	SetObject(d.Object)
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[dbpool] 关闭成功")
	d.Object = nil
}
