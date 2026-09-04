package router

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	obj *Router
}

func (d *Component) Init(data map[string]any) error {
	d.obj = NewRouter()
	SetObject(d.obj)
	mlog.Infof("[router] 初始化成功")
	return nil
}

func (d *Component) Close() {
	mlog.Infof("[router] 关闭成功")
	d.obj = nil
}
