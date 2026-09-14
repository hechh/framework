package infra

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hechh/framework/library/fileutil"
	"github.com/hechh/framework/tools/redistool/domain"
)

// memOutput 收集生成结果（不落盘）
type memOutput struct{ content map[string]string }

func (m *memOutput) Write(filename string, content []byte) error {
	if m.content == nil {
		m.content = make(map[string]string)
	}
	m.content[filename] = string(content)
	return nil
}

func newParser() *ASTParser {
	return NewASTParser(&domain.ParseContext{})
}

// TestBuildStringModel_MissingSegments 验证注解缺段时跳过而不是越界崩溃。
//
// extractModels 只保证 parts[0]（规则类型）存在：写成 // @dbtool:string 或漏掉
// DbSpec/keyFmt 段时，取 parts[1]/parts[2] 会 index out of range 崩栈。
func TestBuildStringModel_MissingSegments(t *testing.T) {
	cases := map[string][]string{
		"只有类型段":      {"@dbtool:string"},
		"缺 keyFmt 段": {"@dbtool:string", "shards:uid@uint64"},
	}
	for name, parts := range cases {
		t.Run(name, func(t *testing.T) {
			if m := newParser().buildStringModel(parts, "TestConfig"); m != nil {
				t.Fatalf("缺段注解必须跳过，实际生成了模型: %+v", m)
			}
		})
	}
}

// TestBuildStringModel_FormatWithoutArgs 验证"格式串含占位符但未声明参数"被拦下。
//
// 该写法得到 format="user_info:%d:%s"、keys=nil；模板 templates.go 的 GetKey 在 Keys
// 为空时直接返回字面量 "{{.Format}}" → 该模型所有 uid 共用同一把 key、互相覆盖
// （串号 + 数据丢失），因此必须报错跳过。
func TestBuildStringModel_FormatWithoutArgs(t *testing.T) {
	parts := []string{"@dbtool:string", "shards:uid@uint64", "user_info:%d:%s"}
	if m := newParser().buildStringModel(parts, "UserInfo"); m != nil {
		t.Fatalf("含占位符但无参数的格式串必须跳过，实际 Format=%q Keys=%v", m.Format, m.Keys)
	}
}

// TestBuildStringModel_StaticKey 验证无占位符的静态 key 仍然合法（不误伤）
func TestBuildStringModel_StaticKey(t *testing.T) {
	m := newParser().buildStringModel([]string{"@dbtool:string", "global:database.REDIS_GLOBAL", "rtp_state"}, "RtpState")
	if m == nil {
		t.Fatal("静态 key 必须合法")
	}
	if m.Format != "rtp_state" || len(m.Keys) != 0 {
		t.Fatalf("静态 key 解析错误: Format=%q Keys=%v", m.Format, m.Keys)
	}
}

// TestBuildStringModel_RealAnnotationForm 验证现有注解使用的 prefix:Field@type 形式
// 照常解析（含多参数），确保新增校验不误伤存量写法
func TestBuildStringModel_RealAnnotationForm(t *testing.T) {
	m := newParser().buildStringModel(
		[]string{"@dbtool:string", "shards:uid@uint64", "user_data:uid@uint64,name@string"}, "UserData")
	if m == nil {
		t.Fatal("合法注解被跳过")
	}
	if m.Format != "user_data:%d:%s" || len(m.Keys) != 2 {
		t.Fatalf("多参数 key 解析错误: Format=%q Keys=%v", m.Format, m.Keys)
	}
	if m.ShardField == nil || m.ShardField.Name != "uid" {
		t.Fatalf("分片字段解析错误: %+v", m.ShardField)
	}
}

// TestHashTemplate_HMGetReturnsUnmarshalError 验证模板生成的 HMGet 不吞 Unmarshal 错误。
//
// 修复前模板里是 `err = reterr; continue` 后接 `return result, nil`：Redis 数据损坏/
// 版本不兼容时所有 hash 模型的 HMGet 返回 (部分 map, nil)，失败字段无声缺失。
func TestHashTemplate_HMGetReturnsUnmarshalError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "test.pb.go")
	body := `package pb

// @dbtool:hash|global:database.REDIS_GLOBAL|user_loader:Uid@uint64|Phone@string
type TestHash struct{}
`
	if err := os.WriteFile(src, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := &domain.ParseContext{}
	if err := fileutil.ParseFiles(NewASTParser(ctx), src); err != nil {
		t.Fatalf("解析 %s 失败: %v", src, err)
	}
	if len(ctx.Hashs) != 1 {
		t.Fatalf("期望解析出 1 个 hash 模型, 实际 %d", len(ctx.Hashs))
	}

	hashTpl, err := BuildHashTemplate()
	if err != nil {
		t.Fatal(err)
	}
	out := &memOutput{}
	if err := domain.NewGenerator(nil, hashTpl).GenerateAll(ctx, out); err != nil {
		t.Fatalf("生成失败: %v", err)
	}
	code := out.content["test_hash/TestHash.gen.go"]
	if code == "" {
		t.Fatalf("未生成代码: keys=%v", out.content)
	}
	if !strings.Contains(code, "return result, err") {
		t.Fatalf("HMGet 必须返回 Unmarshal 错误，实际生成代码:\n%s", code)
	}
}
