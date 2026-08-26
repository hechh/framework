package timer

import (
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Timer
}

// 初始化
func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "timer"); err != nil {
		mlog.Errorf("[timer] 配置加载失败 error:%v", err)
		return err
	}

	// 模块初始化
	if err := d.Object.Init(cfg); err != nil {
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
