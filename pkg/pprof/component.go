package pprof

import (
	"fmt"

	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	obj  *Pprof
	Node *packet.Node
}

func (d *Component) Init(data map[string]any) error {
	// 加载配置
	id := fmt.Sprintf("node%d", d.Node.Id)
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, d.Node.Name, id); err != nil {
		mlog.Errorf("[pprof] 配置加载失败 error:%v", err)
		return err
	}

	// 初始化模块
	d.obj = &Pprof{}
	if err := d.obj.Init(cfg); err != nil {
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
