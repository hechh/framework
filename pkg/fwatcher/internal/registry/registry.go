package registry

import (
	"fmt"
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

func GetFileInfo(sheet string) *parser.FileInfo {
	return files[sheet]
}

// Load 加载配置到内存。
// 初始化时 files 为全量集合（所有注册表必须存在，缺失即报错）；
// 增量热更时 Glob 只返回变更的表，已加载过的表不在集合中则跳过（不重复解析、不触发变更回调）。
func Load(files map[string]*parser.FileInfo) error {
	for sheet, par := range parsers {
		file, ok := files[sheet]
		if !ok {
			// 增量热更：非本次变更的表跳过；首次加载（从未加载过）缺失则报错
			if par.IsLoaded() {
				continue
			}
			return fmt.Errorf("配置文件不存在 sheet:%s", sheet)
		}

		if err := par.Parse(file); err != nil {
			return err
		}
	}
	return nil
}

// 获取所有需要上传的配置
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
