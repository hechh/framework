package timer

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Timer
	Config *Config
}

// 初始化
func (d *Component) Init() error {
	// 模块初始化
	if err := d.Object.Init(d.Config); err != nil {
		mlog.Errorf("[timer] 初始化失败，error=%v", err)
		return err
	}
	mlog.Infof("[timer] 初始化成功")
	SetObject(d.Object)
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[timer] 关闭成功")
	d.Object = nil
}
