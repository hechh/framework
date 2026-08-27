package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hechh/framework/tools/xlsxtool/domain"

	"github.com/xuri/excelize/v2"
)

// ReadTables 从目录读取所有xlsx文件的表格数据。
// 递归遍历时跳过隐藏子目录（如 .bak 暂存、.gen 生成），
// 避免把非正式目录中的 xlsx 当作配置源解析，导致重复表或错误覆盖。
func ReadTables(xlsxDir string) []*domain.Table {
	return ReadTablesOverlay(xlsxDir, "")
}

// ReadTablesOverlay 从 baseDir 读取 xlsx 表格数据，overlayDir 中的同名 xlsx 优先采用。
// 规则：
//   - baseDir 与 overlayDir 均递归遍历，跳过隐藏子目录；
//   - 同一文件名在 overlayDir 存在时，只读取 overlayDir 的版本（覆盖 baseDir）；
//   - baseDir 中 overlayDir 没有的文件，从 baseDir 读取。
//
// 用于"上传暂存覆盖"场景：新上传文件先进 .bak 暂存目录，转换时以 .bak 为准，
// 未上传（.bak 没有）的文件仍从正式目录加载。
func ReadTablesOverlay(baseDir, overlayDir string) []*domain.Table {
	var tables []*domain.Table

	// 1. 收集 overlayDir 中存在的 xlsx 文件名（优先采用）
	overlayNames := make(map[string]bool)
	if overlayDir != "" {
		filepath.Walk(overlayDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if path != overlayDir && strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".xlsx") {
				overlayNames[filepath.Base(path)] = true
			}
			return nil
		})
	}

	// 2. 遍历 baseDir，同名文件在 overlay 中存在时跳过（用 overlay 版本）
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if path != baseDir && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		if overlayNames[filepath.Base(path)] {
			return nil // 由 overlay 提供
		}
		ts, err := readFile(path)
		if err != nil {
			fmt.Printf("[Error] 解析%s失败: %v\n", filepath.Base(path), err)
			return nil
		}
		tables = append(tables, ts...)
		return nil
	})

	// 3. 读取 overlay 中的 xlsx（覆盖/新增）
	for name := range overlayNames {
		ts, err := readFile(filepath.Join(overlayDir, name))
		if err != nil {
			fmt.Printf("[Error] 解析%s失败: %v\n", name, err)
			continue
		}
		tables = append(tables, ts...)
	}
	return tables
}

// ListTypes 读取指定 xlsx 文件“生成表”指令中声明的结构体类型名（对应生成的 XxxConfig.json）。
// 用于“发布指定配置”场景：先全量转换保证枚举/结构引用可解析，再按类型名过滤需要下发的 JSON。
// filenames 中的文件不存在或解析失败时返回错误。
func ListTypes(xlsxDir string, filenames []string) ([]string, error) {
	var types []string
	for _, fn := range filenames {
		path := filepath.Join(xlsxDir, filepath.Base(fn))
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("配置文件不存在: %s", filepath.Base(fn))
		}
		ts, err := readFile(path)
		if err != nil {
			return nil, fmt.Errorf("解析%s失败: %v", filepath.Base(fn), err)
		}
		for _, t := range ts {
			if t.Token == 2 || t.Token == 3 { // @struct / @struct:col 表
				types = append(types, t.Type)
			}
		}
	}
	return types, nil
}

// readFile 读取单个xlsx文件的生成表指令
func readFile(filename string) ([]*domain.Table, error) {
	fp, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	rows, err := fp.GetRows("生成表")
	if err != nil {
		// 无"生成表"指令 sheet：非标准配置表，跳过（由调用方决定是否报错）
		return nil, err
	}

	var tables []*domain.Table
	for _, lines := range rows {
		for _, str := range lines {
			if len(str) <= 0 {
				continue
			}
			vals := strings.Split(str, "|")
			switch {
			case strings.HasPrefix(strings.ToLower(str), "@enum|"):
				sheetRows, err := fp.GetRows(vals[1])
				if err != nil {
					return nil, err
				}
				tables = append(tables, &domain.Table{
					Sheet: vals[1], Rows: sheetRows, Token: 1,
				})
			case strings.HasPrefix(strings.ToLower(str), "@struct|"):
				sheet, typ := vals[1], vals[1]
				if pos := strings.Index(vals[1], "@"); pos >= 0 {
					sheet, typ = vals[1][:pos], vals[1][pos+1:]
				}
				sheetRows, err := fp.GetRows(sheet)
				if err != nil {
					return nil, err
				}
				tables = append(tables, &domain.Table{
					Sheet: sheet, Type: typ,
					Rules: vals[2:], Rows: sheetRows, Token: 2,
				})
			case strings.HasPrefix(strings.ToLower(str), "@struct:col|"):
				sheet, typ := vals[1], vals[1]
				if pos := strings.Index(vals[1], "@"); pos >= 0 {
					sheet, typ = vals[1][:pos], vals[1][pos+1:]
				}
				cols, err := fp.GetCols(sheet)
				if err != nil {
					return nil, err
				}
				tables = append(tables, &domain.Table{
					Sheet: sheet, Type: typ,
					Rules: vals[2:], Rows: cols, Token: 3,
				})
			}
		}
	}
	return tables, nil
}
