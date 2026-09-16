package pprof

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	obj  *Pprof
	Port int32
}

func (d *Component) Init() error {
	// 初始化模块
	d.obj = &Pprof{}
	if err := d.obj.Init(d.Port); err != nil {
		mlog.Errorf("[pprof] 初始化失败，error:%v", err)
		return err
	}
	SetObject(d.obj)
	mlog.Infof("[pprof] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.obj != nil {
		d.obj.Close()
	}
	mlog.Infof("[pprof] 关闭成功")
	d.obj = nil
}
