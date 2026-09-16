package kafka

import (
	"github.com/hechh/framework/pkg/mlog"
)

// Component Kafka 组件：实现 application.IComponent 接口，
// 初始化 Kafka 并自动启动消费循环。
type Component struct {
	Object *Kafka
	Config *Config
}

func (d *Component) Init() error {
	if err := d.Object.Init(d.Config); err != nil {
		mlog.Errorf("[kafka] 初始化失败 error:%v", err)
		return err
	}
	SetObject(d.Object)
	// 自动启动消费循环（订阅主题后即可循环处理消息）
	d.Object.Start()
	mlog.Infof("[kafka] 初始化成功")
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	d.Object = nil
}
