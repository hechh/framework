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

// Register 注册配置解析函数
func RegisterParser[T any](sheet string, parseFunc func(*T) error) {
	registry.Register(sheet, parseFunc)
}

// RegisterChange 注册配置变更回调函数
func RegisterChange(sheet string, changeFunc func()) {
	registry.RegisterChange(sheet, changeFunc)
}
