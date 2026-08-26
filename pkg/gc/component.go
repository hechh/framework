package gc

import "github.com/hechh/framework/pkg/mlog"

type Component struct {
	object *Gc
}

func (d *Component) Init(data map[string]any) error {
	// 初始化模块
	d.object = &Gc{}
	if err := d.object.Init(); err != nil {
		mlog.Errorf("[gc] GC初始化失败，error:%v", err)
		return err
	}
	SetObject(d.object)
	mlog.Infof("[gc] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.object != nil {
		d.object.Close()
	}
	mlog.Infof("[gc] 关闭成功")
	d.object = nil
}
