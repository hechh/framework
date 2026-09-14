package registry

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hechh/framework/pkg/fwatcher/internal/parser"
)

var (
	// filesMu 保护 files：Init 阶段注册与 watch goroutine 的 RegisterFileInfo 会并发写
	// （滚动发布时另一节点推配置即可触发）；Go 的并发 map 写是 fatal error，recover 无效
	filesMu sync.RWMutex
	files   = make(map[string]*parser.FileInfo)
	// parsers 仅在 init() 阶段写入，之后只读，无需加锁
	parsers = make(map[string]parser.IParser)
)

// Register 注册配置解析函数
func Register[T any](sheet string, parseFunc func(*T) error) {
	parsers[sheet] = parser.NewParser(sheet, parseFunc)
}

// RegisterChange 注册配置变更回调函数
func RegisterChange(sheet string, changeFunc func()) {
	if item, ok := parsers[sheet]; ok {
		item.RegisterChange(changeFunc)
	}
}

func GetParser(sheet string) parser.IParser {
	if val, ok := parsers[sheet]; ok {
		return val
	}
	return nil
}

func GetFileInfo(sheet string) *parser.FileInfo {
	filesMu.RLock()
	defer filesMu.RUnlock()
	return files[sheet]
}

func RegisterFileInfo(sheet string, info *parser.FileInfo) {
	filesMu.Lock()
	defer filesMu.Unlock()
	files[sheet] = info
}

func WalkParser(f func(string, parser.IParser) error) error {
	for sheet, par := range parsers {
		if err := f(sheet, par); err != nil {
			return err
		}
	}
	return nil
}

// 获取所有需要上传的配置
func Glob(pattern string) (map[string][]byte, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]byte)
	for _, filename := range matches {
		body, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		sheet := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		result[sheet] = body
	}
	return result, nil
}
