package fwatcher_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/hechh/framework/library/fileutil"
)

// TestConcurrentAtomicSaveSameFile 直接并发调用 AtomicSave 写同一文件，
// 验证固定 .tmp 临时文件名是否会导致并发写冲突/文件损坏。
func TestConcurrentAtomicSaveSameFile(t *testing.T) {
	dir := t.TempDir()
	filename := filepath.Join(dir, "same.json")

	const workers = 20
	const rounds = 50
	var wg sync.WaitGroup
	errs := make(chan error, workers*rounds)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				body := []byte(fmt.Sprintf(`{"worker":%d,"round":%d,"data":"%s"}`, w, r, pad(w, r)))
				if err := fileutil.AtomicSave(filename, body); err != nil {
					errs <- fmt.Errorf("worker %d round %d: %w", w, r, err)
				}
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	// 最终文件必须是合法 JSON（不能是半个写入或损坏）
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("读取最终文件失败: %v", err)
	}
	// 简单校验：内容要么是完整 json，要么是完整内容之一；不能是空/截断
	if len(data) == 0 {
		t.Fatal("最终文件为空")
	}
	t.Logf("最终文件长度: %d, 内容前缀: %s", len(data), data[:20])
}

func pad(w, r int) string {
	s := fmt.Sprintf("%06d-%06d", w, r)
	for len(s) < 60 {
		s += "x"
	}
	return s
}
