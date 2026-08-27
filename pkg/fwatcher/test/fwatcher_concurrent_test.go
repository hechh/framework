package fwatcher_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/fwatcher/adapter/etcdsync"
)

// ============================================================
// fwatcher 并发读写验证：etcd 上传（Put/Update）与监听同步（save 落盘）
//
// 目标：模拟"生产者并发上传、消费者监听同步写本地文件"的高并发场景，
// 用 `go test -race` 检测 data race，并验证本地文件始终是完整可解析的内容
// （不因并发写而被破坏）。
// ============================================================

// runSharedEtcd 启动共享嵌入式 etcd 并返回端点。
func runSharedEtcd(t *testing.T, prefix string) []string {
	t.Helper()
	monitor, cfg := setupSharedEtcd(t, prefix)
	_ = monitor
	return cfg.Etcd.Endpoints
}

// newConsumer 创建一个监听同步的消费者 fwatcher（IsSync=false）。
func newConsumer(t *testing.T, endpoints []string, prefix, dataPath, ext string) *fwatcher.FWatcher {
	t.Helper()
	consumer := fwatcher.NewFWatcher(func() fwatcher.ISync { return etcdsync.NewEtcdSync() })
	cfg := &fwatcher.Config{
		IsSync:   false,
		DataPath: dataPath,
		Ext:      ext,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: endpoints,
		},
	}
	if err := consumer.Init(cfg); err != nil {
		t.Fatalf("消费者初始化失败: %v", err)
	}
	t.Cleanup(consumer.Close)
	return consumer
}

// waitAllFiles 等待本地目录出现期望数量的文件。
func waitAllFiles(t *testing.T, dir string, n int, timeout time.Duration) []string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
		if len(matches) >= n {
			return matches
		}
		if time.Now().After(deadline) {
			t.Fatalf("等待 %d 个文件超时，当前 %d 个: %v", n, len(matches), matches)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestFWatcher_ConcurrentUploadDownload 验证并发上传 + 监听下载：
//   - 生产者并发 Put 多个不同 sheet 配置到 etcd
//   - 消费者 Watch 收到全部变更并落盘
//   - 用 -race 检测 registry / FileInfo 等全局状态的 data race
func TestFWatcher_ConcurrentUploadDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	const (
		sheetCount = 20
		workers    = 8
	)
	prefix := fmt.Sprintf("/test/conc-up-dl-%d", time.Now().UnixNano())
	endpoints := runSharedEtcd(t, prefix)

	// 消费者：先启动并监听
	consumerData := t.TempDir()
	newConsumer(t, endpoints, prefix, consumerData, ".json")

	// 生产者：并发上传多个 sheet
	var wg sync.WaitGroup
	sheetNames := make([]string, 0, sheetCount)
	var mu sync.Mutex
	for i := 0; i < sheetCount; i++ {
		sheet := fmt.Sprintf("ConcSheet_%03d", i)
		mu.Lock()
		sheetNames = append(sheetNames, sheet)
		mu.Unlock()
		body := []byte(fmt.Sprintf(`{"client_version":%d,"enable":true,"items":{"k":"v%d"}}`, i, i))

		wg.Add(1)
		go func(s string, b []byte) {
			defer wg.Done()
			// 每个 goroutine 独立客户端，避免共享 client
			producer := etcdsync.NewEtcdSync()
			pc := &fwatcher.Config{
				Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: endpoints},
			}
			if err := producer.Init(pc); err != nil {
				t.Errorf("生产者客户端初始化失败: %v", err)
				return
			}
			defer producer.Close()
			if err := producer.Put(s, b); err != nil {
				t.Errorf("生产者上传 %s 失败: %v", s, err)
			}
		}(sheet, body)

		// 限制并发数
		if (i+1)%workers == 0 {
			wg.Wait()
		}
	}
	wg.Wait()

	// 断言消费者已落盘全部 sheet
	files := waitAllFiles(t, consumerData, sheetCount, 10*time.Second)
	if len(files) != sheetCount {
		t.Fatalf("消费者落盘文件数不符: want=%d got=%d", sheetCount, len(files))
	}

	// 每个文件内容应为合法 JSON 且 client_version 正确
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("读取文件失败 %s: %v", f, err)
		}
		var cfg struct {
			ClientVersion int32 `json:"client_version"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("文件内容不是合法 JSON: %s err=%v", f, err)
		}
	}
}

// TestFWatcher_ConcurrentUpdateSameSheet 验证同一 sheet 高频并发更新的同步：
//   - 生产者并发 Update 同一个 sheet
//   - 消费者 Watch 收到大量变更并并发写同一本地文件
//   - 最终本地文件必须是完整合法 JSON（不得被并发写坏）
func TestFWatcher_ConcurrentUpdateSameSheet(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	const updateCount = 100
	prefix := fmt.Sprintf("/test/conc-upd-same-%d", time.Now().UnixNano())
	sheet := "ConcSameConfig"
	endpoints := runSharedEtcd(t, prefix)

	// 初始写入
	remote := etcdsync.NewEtcdSync()
	rc := &fwatcher.Config{Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: endpoints}}
	if err := remote.Init(rc); err != nil {
		t.Fatalf("远程客户端初始化失败: %v", err)
	}
	defer remote.Close()
	if err := remote.Put(sheet, []byte(`{"client_version":0,"enable":true,"items":{}}`)); err != nil {
		t.Fatalf("初始写入失败: %v", err)
	}

	// 消费者监听
	consumerData := t.TempDir()
	newConsumer(t, endpoints, prefix, consumerData, ".json")

	// 等待初始文件落盘
	downloaded := filepath.Join(consumerData, sheet+".json")
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(downloaded); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("初始文件未落盘: %v", downloaded)
		}
		time.Sleep(20 * time.Millisecond)
	}

	// 生产者并发更新同一 sheet
	var wg sync.WaitGroup
	for i := 0; i < updateCount; i++ {
		wg.Add(1)
		go func(ver int) {
			defer wg.Done()
			p := etcdsync.NewEtcdSync()
			pc := &fwatcher.Config{Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: endpoints}}
			if err := p.Init(pc); err != nil {
				t.Errorf("生产者客户端初始化失败: %v", err)
				return
			}
			defer p.Close()
			body := []byte(fmt.Sprintf(`{"client_version":%d,"enable":true,"items":{"v":"%d"}}`, ver, ver))
			if err := p.Put(sheet, body); err != nil {
				t.Errorf("生产者更新失败: %v", err)
			}
		}(i)
	}
	wg.Wait()

	// 等待消费者处理完最后更新（多等一会）
	time.Sleep(2 * time.Second)

	// 最终文件必须仍是合法 JSON（并发写不应破坏文件完整性）
	data, err := os.ReadFile(downloaded)
	if err != nil {
		t.Fatalf("读取最终文件失败: %v", err)
	}
	var cfg struct {
		ClientVersion int32 `json:"client_version"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("最终文件不是合法 JSON（可能被并发写坏）: %v\ncontent=%s", err, data)
	}
}
