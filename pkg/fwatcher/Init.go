package fwatcher

import (
	"github.com/hechh/framework/pkg/fwatcher/internal/registry"
)

var (
	object *FWatcher
)

func SetObject(oj *FWatcher) {
	object = oj
}

// GetObject 获取全局 FWatcher 实例（用于模块主动上传/删除配置）。
// 在 fwatcher 组件初始化完成后调用方可用，未初始化时返回 nil。
func GetObject() *FWatcher {
	return object
}

// Register 注册配置解析函数
func RegisterParser[T any](sheet string, parseFunc func(*T) error) {
	registry.Register(sheet, parseFunc)
}

// RegisterChange 注册配置变更回调函数
func RegisterChange(sheet string, changeFunc func()) {
	registry.RegisterChange(sheet, changeFunc)
}
