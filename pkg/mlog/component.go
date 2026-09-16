package mlog

import "fmt"

type Component struct {
	obj    *Logger
	Config *Config
}

func (d *Component) Init() error {
	// 配置缺少 logger 段时 Config 为 nil：此处拦截并给出明确提示，否则 Logger.Init 内部解引用 panic
	if d.Config == nil {
		err := fmt.Errorf("配置为空")
		Errorf("[logger] 初始化失败，error:%v", err)
		return err
	}
	// 初始化模块（obj 未设置时惰性创建）
	d.obj = NewLogger()
	if err := d.obj.Init(d.Config); err != nil {
		Errorf("[logger] 初始化失败，error:%v", err)
		return err
	}
	SetObject(d.obj)
	Infof("[logger] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.obj != nil {
		d.obj.Close()
	}
	Infof("[logger] 关闭成功")
	d.obj = nil
}
