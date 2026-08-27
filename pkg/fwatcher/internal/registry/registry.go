package registry

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hechh/framework/pkg/fwatcher/internal/parser"
)

var (
	files   = make(map[string]*parser.FileInfo)
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
	return files[sheet]
}

func RegisterFileInfo(sheet string, info *parser.FileInfo) {
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
