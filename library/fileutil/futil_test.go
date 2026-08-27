package fileutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestAtomicSaveRenameRetry 验证 AtomicSave 在 rename 遇到 “Access is denied” 后能通过重试恢复。
//
// Windows 上目标文件只读（或正被其他进程瞬时占用）会导致 rename 覆盖失败并报
// “Access is denied”（对应多服务共享同一 data 目录并发写同一配置的场景）。
// 延迟恢复可写后，重试应成功且不残留临时文件。
// 其他平台（POSIX）rename 不受目标只读影响，直接成功，测试同样通过。
func TestAtomicSaveRenameRetry(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "test.json")

	// 先写旧内容并设为只读（Windows 上 rename 覆盖只读目标 → Access is denied）
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("写旧文件失败: %v", err)
	}
	if err := os.Chmod(dst, 0o444); err != nil {
		t.Fatalf("设置只读失败: %v", err)
	}

	// 延迟恢复可写，让首次 rename 失败后在重试中成功
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = os.Chmod(dst, 0o644)
	}()

	if err := AtomicSave(dst, []byte("new-content")); err != nil {
		t.Fatalf("AtomicSave 重试后仍失败: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("读取结果失败: %v", err)
	}
	if string(data) != "new-content" {
		t.Fatalf("内容=%q, 期望 new-content", string(data))
	}

	// 不应残留临时文件（成功路径已由 defer 清理）
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "test.json" {
			t.Fatalf("目录存在多余文件: %s", e.Name())
		}
	}
}
