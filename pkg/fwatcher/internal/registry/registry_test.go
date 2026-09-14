package registry

import (
	"fmt"
	"sync"
	"testing"

	"github.com/hechh/framework/pkg/fwatcher/internal/parser"
)

// TestRegisterFileInfo_ConcurrentWrite 验证 files 的并发读写安全。
//
// 场景：Init 阶段（主 goroutine）注册本地已下载的配置文件，同时 watch goroutine
// 收到另一节点推送的配置后也会 RegisterFileInfo。修复前该 map 无任何同步，
// Go 的并发 map 写是 fatal error（recover 无效，直接终止进程），滚动发布时即可触发。
//
// 用 `go test -race` 运行可同时检出该 data race。
func TestRegisterFileInfo_ConcurrentWrite(t *testing.T) {
	const (
		rounds  = 200
		workers = 8
	)

	var wg sync.WaitGroup
	// 并发写（模拟 Init 与 watch goroutine 同时注册）
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				sheet := fmt.Sprintf("Sheet_%d_%d", w, i)
				RegisterFileInfo(sheet, parser.NewFileInfo(sheet, []byte("{}")))
			}
		}(w)
	}
	// 并发读
	for r := 0; r < workers; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < rounds; i++ {
				GetFileInfo(fmt.Sprintf("Sheet_0_%d", i))
			}
		}()
	}
	wg.Wait()

	// 注册结果必须可读回
	info := GetFileInfo("Sheet_0_0")
	if info == nil {
		t.Fatalf("注册后的 FileInfo 读不到")
	}
	if info.GetSheet() != "Sheet_0_0" {
		t.Fatalf("FileInfo 内容错乱: %s", info.GetSheet())
	}
}
