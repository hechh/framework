package parser

import (
	"encoding/json"
	"sync/atomic"

	"github.com/hechh/framework/pkg/mlog"
	"google.golang.org/protobuf/proto"
)

// 配置解析接口
type IParser interface {
	RegisterChange(...func())
	Sheet() string
	GetValue() string
	New([]byte) (proto.Message, error)
	Parse(*FileInfo) error
	IsLoaded() bool // 是否已完成首次加载
}

// Parser 配置解析器
type Parser[T any] struct {
	sheetName   string         // sheet名字
	info        *FileInfo      // 文件信息
	status      atomic.Int32   // 配置解析器状态，0=未加载，1=已加载
	parseFunc   func(*T) error // 配置解析函数
	changeFuncs []func()       // 变更回调函数列表
}

func NewParser[T any](sheet string, f func(*T) error) *Parser[T] {
	return &Parser[T]{
		sheetName: sheet,
		parseFunc: f,
	}
}

// RegisterChange 注册配置变更回调函数
func (p *Parser[T]) RegisterChange(callbacks ...func()) {
	p.changeFuncs = append(p.changeFuncs, callbacks...)
}

// Sheet 获取配置表名
func (p *Parser[T]) Sheet() string {
	return p.sheetName
}

func (p *Parser[T]) GetValue() string {
	if p.info != nil {
		return p.info.GetValue()
	}
	return ""
}

// IsLoaded 是否已完成首次加载（status: 0=未加载，1=已加载）
func (p *Parser[T]) IsLoaded() bool {
	return p.status.Load() == 1
}

func (p *Parser[T]) New(body []byte) (proto.Message, error) {
	val := any(new(T)).(proto.Message)
	if err := json.Unmarshal(body, val); err != nil {
		return nil, err
	}
	return val, nil
}

// Parse 解析配置内容
func (p *Parser[T]) Parse(file *FileInfo) error {
	p.info = file

	// 解析配置内容（JSON，键名与 pb.go json tag 一致）
	ary := new(T)
	if err := json.Unmarshal(file.GetText(), ary); err != nil {
		return err
	}
	// 加载配置（如果已注册解析函数）
	if err := p.parseFunc(ary); err != nil {
		return err
	}
	mlog.Tracef("成功加载配置:%s", p.sheetName)

	// 配置变更通知（仅在非初始化加载时触发）
	if !p.status.CompareAndSwap(0, 1) {
		for _, f := range p.changeFuncs {
			f()
		}
	}
	return nil
}
