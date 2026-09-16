package redispool

import (
	"github.com/hechh/framework/pkg/mlog"
)

type Component struct {
	Object  *RedisPool
	Configs map[string][]*Config
}

// 初始化
func (d *Component) Init() error {
	// 模块初始化
	if err := d.Object.Init(d.Configs["globals"], d.Configs["shards"]); err != nil {
		mlog.Errorf("[redispool] 初始化失败，error=%v", err)
		return err
	}
	mlog.Infof("[redispool] 初始化成功")
	SetObject(d.Object)
	return nil
}

func (d *Component) Close() {
	if d.Object != nil {
		d.Object.Close()
	}
	mlog.Infof("[redispool] 关闭成功")
	d.Object = nil
}
