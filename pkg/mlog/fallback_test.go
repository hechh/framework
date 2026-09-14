package mlog

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRotateWriter_FallbackToStderrOnWriteFailure 验证日志文件写失败时降级到 stderr 并告警。
//
// 修复前：handler 里 cache.Set/cache.Write 的 error 全部被丢弃，release 模式又只挂
// RotateWriter（没有 stdout），日志目录不可写时日志全部丢失且零提示，事故时无日志可查。
func TestRotateWriter_FallbackToStderrOnWriteFailure(t *testing.T) {
	dir := t.TempDir()
	// 用普通文件占住日志目录路径：CreateFile 内的 MkdirAll 必然失败，稳定复现"目录不可写"
	blocker := filepath.Join(dir, "log")
	if err := os.WriteFile(blocker, []byte("not a dir"), 0o644); err != nil {
		t.Fatalf("准备阻塞文件失败: %v", err)
	}

	// 抓取 stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("创建管道失败: %v", err)
	}
	os.Stderr = w

	writer := NewRotateWriter(blocker, "fallback_test", 1024, 10*time.Millisecond, RollingByDay)
	writer.Write(time.Now(), []byte("fallback-log-content\n"))
	time.Sleep(200 * time.Millisecond)
	_ = writer.Close()

	w.Close()
	os.Stderr = oldStderr
	out, _ := io.ReadAll(r)

	got := string(out)
	if !strings.Contains(got, "fallback-log-content") {
		t.Fatalf("文件写入失败时日志必须降级输出到 stderr, got=%q", got)
	}
	if !strings.Contains(got, "日志写入文件失败") {
		t.Fatalf("文件写入失败必须告警（否则日志静默丢失）, got=%q", got)
	}
}
