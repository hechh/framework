package mlog

type Component struct {
	obj    *Logger
	Config *Config
}

func (d *Component) Init() error {
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
