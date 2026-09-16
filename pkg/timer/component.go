package timer

import (
	"fmt"

	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object *Timer
	Config *Config
}

// 初始化
func (d *Component) Init() error {
	// 配置缺少 timer 段时 Config 为 nil：此处拦截并给出明确提示，否则 Timer.Init 内部解引用 panic
	if d.Config == nil {
		err := fmt.Errorf("配置为空")
		mlog.Errorf("[timer] 初始化失败，error=%v", err)
		return err
	}
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
