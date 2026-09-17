package infra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

// writeTestTable 生成单表测试 xlsx（生成表指向 Data@<typ>，含 字段名/类型/说明 三行表头与一行数据）。
func writeTestTable(t *testing.T, path, typ string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	f := excelize.NewFile()
	genIdx, _ := f.NewSheet("生成表")
	f.SetActiveSheet(genIdx)
	f.SetCellValue("生成表", "A1", "@struct|Data@"+typ)
	f.NewSheet("Data")
	f.SetCellValue("Data", "A1", "Id")
	f.SetCellValue("Data", "B1", "Name")
	f.SetCellValue("Data", "A2", "int32")
	f.SetCellValue("Data", "B2", "string")
	f.SetCellValue("Data", "A3", "id")
	f.SetCellValue("Data", "B3", "name")
	f.SetCellValue("Data", "A4", "1")
	f.SetCellValue("Data", "B4", "x")
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("保存测试 xlsx 失败: %v", err)
	}
}

// TestReadTablesEx_SkipDirs 验证 ReadTablesEx 跳过指定子目录（如配置管理的 waiting 上传暂存目录）：
// 1. 正式目录的表被读取；
// 2. skipDirs 指定的 waiting 目录中的表不被读取（即使它不是隐藏目录）；
// 3. 隐藏目录（.bak）仍按既有语义跳过；
// 4. ReadTables 保持原语义（waiting 会被读取，只有隐藏目录被跳过）。
func TestReadTablesEx_SkipDirs(t *testing.T) {
	dir := t.TempDir()
	writeTestTable(t, filepath.Join(dir, "a.xlsx"), "AType")
	writeTestTable(t, filepath.Join(dir, "waiting", "b.xlsx"), "BType")
	writeTestTable(t, filepath.Join(dir, ".bak", "c.xlsx"), "CType")

	got := make(map[string]bool)
	for _, tb := range ReadTablesEx(dir, "waiting") {
		if tb.Type != "" {
			got[tb.Type] = true
		}
	}
	if !got["AType"] {
		t.Errorf("缺少正式目录的表 AType: %v", got)
	}
	if got["BType"] {
		t.Errorf("skipDirs 指定的 waiting 目录不应被读取: %v", got)
	}
	if got["CType"] {
		t.Errorf("隐藏目录 .bak 不应被读取: %v", got)
	}

	// ReadTables 保持原语义：仅跳过隐藏目录
	got2 := make(map[string]bool)
	for _, tb := range ReadTables(dir) {
		if tb.Type != "" {
			got2[tb.Type] = true
		}
	}
	if !got2["AType"] || !got2["BType"] {
		t.Errorf("ReadTables 应保持原语义（waiting 可被读取）: %v", got2)
	}
	if got2["CType"] {
		t.Errorf("ReadTables 仍应跳过隐藏目录: %v", got2)
	}
}
