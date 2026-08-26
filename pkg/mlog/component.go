package mlog

import (
	"fmt"

	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/library/tplutil"
	"github.com/hechh/framework/packet"
)

type Component struct {
	obj  *Logger
	Node *packet.Node
}

func (d *Component) Init(data map[string]any) error {
	// 加载服务节点配置
	id := fmt.Sprintf("node%d", d.Node.Id)
	tmpCfg := &Config{}
	fileutil.Map2Yaml(data, tmpCfg, d.Node.Name, id, "logger")

	// 加载配置
	cfg := &Config{}
	if err := fileutil.Map2Yaml(data, cfg, "logger"); err != nil {
		Errorf("[logger] 配置加载失败 error:%v", err)
		return err
	}

	cfg.Name = fmt.Sprintf("%s%d", d.Node.Name, d.Node.Id)
	cfg.Mode = tplutil.Or(tmpCfg.Mode == "", cfg.Mode, tmpCfg.Mode)
	cfg.Path = tplutil.Or(tmpCfg.Path == "", cfg.Path, tmpCfg.Path)
	cfg.Level = tplutil.Or(tmpCfg.Level == "", cfg.Level, tmpCfg.Level)
	cfg.Format = tplutil.Or(tmpCfg.Format == "", cfg.Format, tmpCfg.Format)
	cfg.IsCaller = tplutil.Or(tmpCfg.IsCaller, tmpCfg.IsCaller, cfg.IsCaller)
	cfg.CacheSize = tplutil.Or(tmpCfg.CacheSize == 0, cfg.CacheSize, tmpCfg.CacheSize)

	// 初始化模块（obj 未设置时惰性创建）
	d.obj = NewLogger()
	if err := d.obj.Init(cfg); err != nil {
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
