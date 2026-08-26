package dbpool

import (
	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *DbPool
}

// 初始化
func (d *Component) Init(data map[string]any) error {
	// 加载配置
	cfg := make(map[string]*Config)
	if err := fileutil.Map2Yaml(data, cfg, "dbpool"); err != nil {
		mlog.Errorf("[dbpool] 配置加载失败 error:%v", err)
		return err
	}

	// 模块初始化
	if err := d.Object.Init(cfg); err != nil {
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
