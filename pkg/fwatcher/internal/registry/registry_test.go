package registry

import (
	"os"
	"path/filepath"
	"testing"
)

// testCfg 用于测试的简化配置结构（仅 JSON 反序列化）
type testCfg struct {
	Name string `json:"name"`
}

// TestLoadIncremental 验证增量热更：单个配置变更时 Load 不应报"配置文件不存在"，
// 且只重新解析变更的表、只对变更的表再次调用 parseFunc。
//
// 背景：adminsrv 上传 xlsx 后，其他服务通过 etcd → save 逐个写入本地 json，
// 每次本地文件变更触发 watch() → Glob() 只返回变更的文件 → Load(变更集)。
// 修复前 Load 要求所有注册表都在集合中，导致 "配置文件不存在 sheet:Xxx"。
func TestLoadIncremental(t *testing.T) {
	// 临时数据目录，三个配置文件
	dir := t.TempDir()
	sheets := []string{"Alpha", "Beta", "Gamma"}
	for _, name := range sheets {
		if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(`{"name":"`+name+`"}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// 注册解析器（模拟 xlsxtool 生成的 init()），统计每个表被解析的次数
	counts := map[string]int{}
	for _, name := range sheets {
		sheet := name
		Register[testCfg](sheet, func(c *testCfg) error {
			counts[c.Name]++
			return nil
		})
	}

	// ① 首次全量加载（模拟 FWatcher.Init：缓存为空，Glob 返回全部）
	files, err := Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 3 {
		t.Fatalf("首次 Glob 应返回 3 个文件, got %d", len(files))
	}
	if err := Load(files); err != nil {
		t.Fatalf("首次 Load 失败: %v", err)
	}
	for _, name := range sheets {
		if counts[name] != 1 {
			t.Fatalf("首次加载应解析 %s 一次, got %d", name, counts[name])
		}
	}

	// ② 模拟 adminsrv 上传导致单个配置变更（etcd→save→本地文件变更→watch→Glob→Load）
	if err := os.WriteFile(filepath.Join(dir, "Beta.json"), []byte(`{"name":"Beta2"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err = Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("增量 Glob 应只返回变更的 1 个文件, got %d", len(files))
	}
	// 修复前这里会报 "配置文件不存在 sheet:xxx"
	if err := Load(files); err != nil {
		t.Fatalf("增量 Load 不应报错: %v", err)
	}
	// 只有 Beta 被重新解析，Alpha/Gamma 保持不变
	if counts["Alpha"] != 1 || counts["Gamma"] != 1 {
		t.Fatalf("未变更的表不应被重新解析: %v", counts)
	}
	if counts["Beta2"] != 1 {
		t.Fatalf("变更的 Beta 应被重新解析一次: %v", counts)
	}
}
