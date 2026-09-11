package fwatcher_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/fwatcher/adapter/etcdsync"
	"github.com/hechh/framework/pkg/fwatcher/adapter/mocksync"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// ============================================================
// fwatcher 集成测试：验证配置上传（Put/全量同步）与下载（Fetch/Watch）
//
// 说明：
//   - 依赖嵌入式 etcd（mocksync.EmbedSync），会真实启动一个 etcd 服务。
//   - 用 `go test -short` 可跳过（适合无嵌入式 etcd 依赖的 CI 环境）。
//   - 由于 fwatcher 内部 registry 为包级全局状态，各测试使用独立 sheet 名
//     隔离；上传/下载分别用单实例 fwatcher + 独立 etcdsync 客户端验证，
//     避免同进程多实例相互污染（真实生产环境一进程一 fwatcher）。
// ============================================================

// testConfig 用于注册 parser 的测试配置结构
type testConfig struct {
	ClientVersion int32             `json:"client_version"`
	Enable        bool              `json:"enable"`
	Items         map[string]string `json:"items"`
}

// setupSharedEtcd 启动一个共享的嵌入式 etcd。
// 返回的 cfg.Etcd.Endpoints 已被改写为嵌入式 etcd 的真实地址，供客户端连接。
func setupSharedEtcd(t *testing.T, prefix string) (*mocksync.EmbedSync, *fwatcher.Config) {
	t.Helper()

	cfg := &fwatcher.Config{
		IsSync: false,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: []string{"http://127.0.0.1:2379"}, // 占位，Init 后会被覆盖为嵌入式地址
		},
	}

	monitor := mocksync.NewMonitor()
	if err := monitor.Init(cfg); err != nil {
		t.Fatalf("启动嵌入式 etcd 失败: %v", err)
	}
	t.Cleanup(monitor.Close)
	return monitor, cfg
}

// newEtcdClient 直连嵌入式 etcd，用于断言上传/下载的实际数据。
func newEtcdClient(t *testing.T, endpoints []string) *clientv3.Client {
	t.Helper()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("创建 etcd 客户端失败: %v", err)
	}
	t.Cleanup(func() { cli.Close() })
	return cli
}

// getEtcdValue 读取 etcd 中指定 key 的值。
func getEtcdValue(t *testing.T, cli *clientv3.Client, key string) ([]byte, bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rsp, err := cli.Get(ctx, key)
	if err != nil {
		t.Fatalf("读取 etcd key(%s) 失败: %v", key, err)
	}
	if len(rsp.Kvs) == 0 {
		return nil, false
	}
	return rsp.Kvs[0].Value, true
}

// listEtcdPrefix 列出 etcd 中某前缀下的所有 key。
func listEtcdPrefix(t *testing.T, cli *clientv3.Client, prefix string) []string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rsp, err := cli.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		t.Fatalf("列出 etcd 前缀(%s) 失败: %v", prefix, err)
	}
	keys := make([]string, 0, len(rsp.Kvs))
	for _, kv := range rsp.Kvs {
		keys = append(keys, string(kv.Key))
	}
	return keys
}

// writeDataFile 向 dataPath 写一个配置文件，返回完整文件名。
func writeDataFile(t *testing.T, dataPath, sheet, ext string, body []byte) string {
	t.Helper()
	if err := os.MkdirAll(dataPath, 0o755); err != nil {
		t.Fatalf("创建数据目录失败: %v", err)
	}
	filename := filepath.Join(dataPath, sheet+ext)
	if err := os.WriteFile(filename, body, 0o644); err != nil {
		t.Fatalf("写入数据文件失败: %v", err)
	}
	return filename
}

// registerTestParser 注册一个测试 parser，返回一个函数用于读取最近一次加载的内容。
func registerTestParser(t *testing.T, sheet string) (loaded func() *testConfig) {
	t.Helper()
	var (
		mu  sync.Mutex
		cur *testConfig
	)
	fwatcher.RegisterParser(sheet, func(c *testConfig) error {
		mu.Lock()
		defer mu.Unlock()
		cur = c
		return nil
	})
	return func() *testConfig {
		mu.Lock()
		defer mu.Unlock()
		return cur
	}
}

// waitFile 轮询等待本地文件出现，返回其内容。
func waitFile(t *testing.T, filename string, want []byte, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		data, err := os.ReadFile(filename)
		if err == nil && (want == nil || string(data) == string(want)) {
			return data
		}
		if time.Now().After(deadline) {
			data, _ := os.ReadFile(filename)
			t.Fatalf("等待文件超时 %s：\nwant=%s\ngot =%s", filename, want, data)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ============================================================
// 上传：验证 fwatcher(IsSync=true) 初始化时把本地配置全量上传到 etcd
// ============================================================
func TestFWatcher_UploadConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	prefix := fmt.Sprintf("/test/upload-%d", time.Now().UnixNano())
	sheet := "UploadGlobalConfig"
	ext := ".json"
	body := []byte(`{"client_version":1,"enable":true,"items":{"a":"1","b":"2"}}`)

	_, cfg := setupSharedEtcd(t, prefix)

	// 发布者（IsSync=true）的上传源是 xlsx 生成目录 <XlsxPath>/json，消费者缓存目录 DataPath 另置，
	// 两者分离可避免消费者的 etcd 落盘覆盖发布者的真源快照（见 fwatcher.Init）。
	xlsxPath := t.TempDir()
	genPath := filepath.Join(xlsxPath, "json")
	writeDataFile(t, genPath, sheet, ext, body)
	sheet2 := "UploadExtraConfig"
	body2 := []byte(`{"client_version":2,"enable":false,"items":{}}`)
	writeDataFile(t, genPath, sheet2, ext, body2)

	// 生产者 fwatcher：IsSync=true，Init 时把生成目录中的配置全量上传
	producer := fwatcher.NewFWatcher(func() fwatcher.ISync { return etcdsync.NewEtcdSync() })
	producerCfg := &fwatcher.Config{
		IsSync:   true,
		XlsxPath: xlsxPath,
		DataPath: t.TempDir(),
		Ext:      ext,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: cfg.Etcd.Endpoints,
		},
	}
	if err := producer.Init(producerCfg); err != nil {
		t.Fatalf("生产者初始化失败: %v", err)
	}
	defer producer.Close()

	// 断言 etcd 中两个 key 都存在且内容正确
	cli := newEtcdClient(t, cfg.Etcd.Endpoints)
	if got, ok := getEtcdValue(t, cli, prefix+"/"+sheet); !ok {
		t.Fatalf("上传失败：etcd 中不存在 key(%s)", prefix+"/"+sheet)
	} else if string(got) != string(body) {
		t.Fatalf("上传内容不一致：\nwant=%s\ngot =%s", body, got)
	}
	if got, ok := getEtcdValue(t, cli, prefix+"/"+sheet2); !ok {
		t.Fatalf("上传失败：etcd 中不存在 key(%s)", prefix+"/"+sheet2)
	} else if string(got) != string(body2) {
		t.Fatalf("上传内容不一致：\nwant=%s\ngot =%s", body2, got)
	}
}

// ============================================================
// 下载：验证 fwatcher(IsSync=false) 初始化时从 etcd 拉取全量配置到本地
// ============================================================
func TestFWatcher_DownloadConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	prefix := fmt.Sprintf("/test/download-%d", time.Now().UnixNano())
	sheet := "DownloadGlobalConfig"
	ext := ".json"
	body := []byte(`{"client_version":5,"enable":true,"items":{"k":"v"}}`)

	_, cfg := setupSharedEtcd(t, prefix)

	// 用独立 etcdsync 客户端预先写入 etcd（模拟远端已有配置）
	remote := etcdsync.NewEtcdSync()
	remoteCfg := &fwatcher.Config{
		Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: cfg.Etcd.Endpoints},
	}
	if err := remote.Init(remoteCfg); err != nil {
		t.Fatalf("远程客户端初始化失败: %v", err)
	}
	defer remote.Close()
	if err := remote.Put(sheet, body); err != nil {
		t.Fatalf("写入 etcd 配置失败: %v", err)
	}

	// 消费者 fwatcher：IsSync=false，从空目录开始，Init 时经 Watch/Fetch 下载
	loaded := registerTestParser(t, sheet)
	consumerData := t.TempDir()
	consumer := fwatcher.NewFWatcher(func() fwatcher.ISync { return etcdsync.NewEtcdSync() })
	consumerCfg := &fwatcher.Config{
		IsSync:   false,
		DataPath: consumerData,
		Ext:      ext,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: cfg.Etcd.Endpoints,
		},
	}
	if err := consumer.Init(consumerCfg); err != nil {
		t.Fatalf("消费者初始化失败: %v", err)
	}
	defer consumer.Close()

	// 断言本地文件已生成且内容一致
	downloaded := filepath.Join(consumerData, sheet+ext)
	got := waitFile(t, downloaded, body, 5*time.Second)
	if string(got) != string(body) {
		t.Fatalf("下载内容不一致：\nwant=%s\ngot =%s", body, got)
	}

	// 断言内存中 parser 已加载（parseFunc 被调用）
	deadline := time.Now().Add(5 * time.Second)
	for {
		if cur := loaded(); cur != nil && cur.ClientVersion == 5 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("消费者内存中未加载到配置：ClientVersion != 5")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ============================================================
// 变更同步：验证 fwatcher 通过 Watch 实时接收 etcd 配置更新并落盘
// ============================================================
func TestFWatcher_UpdateSync(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	prefix := fmt.Sprintf("/test/update-%d", time.Now().UnixNano())
	sheet := "UpdateGlobalConfig"
	ext := ".json"
	body := []byte(`{"client_version":1,"enable":true,"items":{}}`)

	_, cfg := setupSharedEtcd(t, prefix)

	// 先用独立客户端写入初始配置
	remote := etcdsync.NewEtcdSync()
	remoteCfg := &fwatcher.Config{
		Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: cfg.Etcd.Endpoints},
	}
	if err := remote.Init(remoteCfg); err != nil {
		t.Fatalf("远程客户端初始化失败: %v", err)
	}
	defer remote.Close()
	if err := remote.Put(sheet, body); err != nil {
		t.Fatalf("写入 etcd 配置失败: %v", err)
	}

	// 消费者：IsSync=false，监听 etcd 变更
	loaded := registerTestParser(t, sheet)
	consumerData := t.TempDir()
	consumer := fwatcher.NewFWatcher(func() fwatcher.ISync { return etcdsync.NewEtcdSync() })
	consumerCfg := &fwatcher.Config{
		IsSync:   false,
		DataPath: consumerData,
		Ext:      ext,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: cfg.Etcd.Endpoints,
		},
	}
	if err := consumer.Init(consumerCfg); err != nil {
		t.Fatalf("消费者初始化失败: %v", err)
	}
	defer consumer.Close()

	// 先确认初始配置已下载
	downloaded := filepath.Join(consumerData, sheet+ext)
	waitFile(t, downloaded, body, 5*time.Second)

	// 远端更新配置 → 消费者 Watch 应实时收到并写盘
	newBody := []byte(`{"client_version":2,"enable":false,"items":{"x":"y"}}`)
	if err := remote.Put(sheet, newBody); err != nil {
		t.Fatalf("远端更新配置失败: %v", err)
	}

	got := waitFile(t, downloaded, newBody, 5*time.Second)
	if string(got) != string(newBody) {
		t.Fatalf("变更同步内容不一致：\nwant=%s\ngot =%s", newBody, got)
	}

	// 断言内存已刷新
	deadline := time.Now().Add(5 * time.Second)
	for {
		if cur := loaded(); cur != nil && cur.ClientVersion == 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("消费者内存中未更新到新配置：ClientVersion != 2")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// ============================================================
// 删除：验证 etcd 配置删除后 key 消失（本地文件归档为 .deleted）
// ============================================================
func TestFWatcher_DeleteSync(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	prefix := fmt.Sprintf("/test/delete-%d", time.Now().UnixNano())
	sheet := "DeleteGlobalConfig"
	ext := ".json"
	body := []byte(`{"client_version":1,"enable":true,"items":{}}`)

	_, cfg := setupSharedEtcd(t, prefix)

	// 先写入初始配置
	remote := etcdsync.NewEtcdSync()
	remoteCfg := &fwatcher.Config{
		Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: cfg.Etcd.Endpoints},
	}
	if err := remote.Init(remoteCfg); err != nil {
		t.Fatalf("远程客户端初始化失败: %v", err)
	}
	defer remote.Close()
	if err := remote.Put(sheet, body); err != nil {
		t.Fatalf("写入 etcd 配置失败: %v", err)
	}

	// 消费者：IsSync=false，监听变更
	consumerData := t.TempDir()
	consumer := fwatcher.NewFWatcher(func() fwatcher.ISync { return etcdsync.NewEtcdSync() })
	consumerCfg := &fwatcher.Config{
		IsSync:   false,
		DataPath: consumerData,
		Ext:      ext,
		Etcd: &fwatcher.EtcdConfig{
			Prefix:    prefix,
			Endpoints: cfg.Etcd.Endpoints,
		},
	}
	if err := consumer.Init(consumerCfg); err != nil {
		t.Fatalf("消费者初始化失败: %v", err)
	}
	defer consumer.Close()

	// 先确认初始配置已下载
	downloaded := filepath.Join(consumerData, sheet+ext)
	waitFile(t, downloaded, body, 5*time.Second)

	// 远端删除配置
	if err := remote.Delete(sheet); err != nil {
		t.Fatalf("远端删除配置失败: %v", err)
	}

	// 断言 etcd 中 key 已删除
	cli := newEtcdClient(t, cfg.Etcd.Endpoints)
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, ok := getEtcdValue(t, cli, prefix+"/"+sheet); !ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("etcd 中 key(%s) 未被删除", prefix+"/"+sheet)
		}
		time.Sleep(50 * time.Millisecond)
	}

	// 消费者本地文件应归档为 .deleted（删除事件 body==nil，不覆盖原文件，重命名保留备份）
	deleted := downloaded + ".deleted"
	if _, err := os.Stat(deleted); err != nil {
		t.Fatalf("删除事件应将本地文件归档为 .deleted: %v", err)
	}
	if _, err := os.Stat(downloaded); err == nil {
		t.Fatalf("删除事件后原文件应被重命名为 .deleted")
	}
}

// ============================================================
// 按 revision 续传：验证「快照与建流之间」以及「断线期间」的变更不会丢失
// ============================================================
//
// etcd watch 不带 WithRev 时只收建流之后的新事件。若不记录 revision 续传，
// Fetch 快照与建流之间（重连时则是断线期间）发生的 Put/Delete 会被静默丢弃，
// 表现为「运营改了配置，部分实例长期不生效」。本用例固定该行为。
func TestEtcdSync_ResumeFromRevision(t *testing.T) {
	if testing.Short() {
		t.Skip("集成测试需要嵌入式 etcd，-short 模式下跳过")
	}

	prefix := fmt.Sprintf("/test/resume-%d", time.Now().UnixNano())
	_, cfg := setupSharedEtcd(t, prefix)

	remote := etcdsync.NewEtcdSync()
	if err := remote.Init(&fwatcher.Config{
		Etcd: &fwatcher.EtcdConfig{Prefix: prefix, Endpoints: cfg.Etcd.Endpoints},
	}); err != nil {
		t.Fatalf("EtcdSync 初始化失败: %v", err)
	}
	defer remote.Close()

	// 1. 建立基线快照（此时 prefix 下无配置）
	snapshot := make(map[string]string)
	if err := remote.Fetch(func(key string, body []byte) { snapshot[key] = string(body) }); err != nil {
		t.Fatalf("Fetch 失败: %v", err)
	}
	if len(snapshot) != 0 {
		t.Fatalf("基线快照应为空，实际 %v", snapshot)
	}

	// 2. 快照之后、建流之前写入配置（等价于断线期间发生的变更）
	sheet := "ResumeGlobalConfig"
	body := []byte(`{"client_version":9,"enable":true,"items":{"k":"v"}}`)
	if err := remote.Put(sheet, body); err != nil {
		t.Fatalf("Put 失败: %v", err)
	}

	// 3. 再写入一次无关配置，把 etcd 当前 revision 推后：
	//    若只写 2 中这一笔，watch(rev=0) 恰好会带上「最近一次提交」而掩盖缺陷，
	//    推后 revision 后才能真实区分「续传」与「只收建流之后的新事件」。
	if err := remote.Put("ResumeFillerConfig", []byte(`{"client_version":1,"enable":true,"items":{}}`)); err != nil {
		t.Fatalf("写入无关配置失败: %v", err)
	}

	// 4. 开始监听：应从快照 revision 之后续传，立即收到 2 中的变更
	got := make(chan string, 1)
	if err := remote.Watch(func(key string, value []byte) {
		if key == prefix+"/"+sheet {
			select {
			case got <- string(value):
			default:
			}
		}
	}); err != nil {
		t.Fatalf("Watch 失败: %v", err)
	}

	select {
	case v := <-got:
		if v != string(body) {
			t.Fatalf("续传内容不一致: want=%s got=%s", body, v)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("续传失败：快照之后、建流之前写入的配置未被重放（断线期间的变更会静默丢失）")
	}
}
