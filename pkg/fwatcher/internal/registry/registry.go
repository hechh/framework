package registry

import (
	"fmt"

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

func GetFileInfo(sheet string) *parser.FileInfo {
	return files[sheet]
}

func GetParser(sheet string) parser.IParser {
	if val, ok := parsers[sheet]; ok {
		return val
	}
	return nil
}

func RegisterFileInfo(sheet string, info *parser.FileInfo) {
	files[sheet] = info
}

// Load 加载配置到内存。
func Load(files map[string]*parser.FileInfo) error {
	for sheet, par := range parsers {
		file, ok := files[sheet]
		if !ok {
			return fmt.Errorf("配置文件不存在 sheet:%s", sheet)
		}
		if err := par.Parse(file); err != nil {
			return err
		}
	}
	return nil
}

// 获取所有需要上传的配置
/*
func Glob(pattern string) (map[string]*parser.FileInfo, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*parser.FileInfo)
	for _, filename := range matches {
		body, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		sheet := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		if item, ok := files[sheet]; !ok {
			item := parser.NewFileInfo(filename, body)
			files[sheet] = item
			result[sheet] = item
		} else if item.Update(body) {
			result[sheet] = item
		}
	}
	return result, nil
}
*/
