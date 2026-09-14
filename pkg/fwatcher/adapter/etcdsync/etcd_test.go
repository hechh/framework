package etcdsync

import (
	"testing"

	"github.com/hechh/framework/pkg/fwatcher"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// TestEtcdSyncClose_Idempotent 验证重复 Close 不 panic。
//
// 与 cluster/adapter/discovery 同构：fwatcher 组件的初始化失败路径与组件 Close
// 都可能调用 Close，修复前第二次 close(exitCh) 会 panic。
// 依赖本机 etcd（127.0.0.1:12379），不可用时跳过。
func TestEtcdSyncClose_Idempotent(t *testing.T) {
	d := NewEtcdSync()
	cfg := &fwatcher.Config{Etcd: &fwatcher.EtcdConfig{
		Prefix:    "test/close-idempotent",
		Endpoints: []string{"127.0.0.1:12379"},
	}}
	if err := d.Init(cfg); err != nil {
		t.Skipf("etcd 不可用，跳过该集成用例: %v", err)
	}

	d.Close()
	d.Close() // 修复前此处 panic
}

// TestReconnect_FailureKeepsChannel 验证重连失败时不会把 watchCh 置空。
//
// 修复前：`watchCh, watchErr = e.watch(f)` 在失败时把 watchCh 赋成 nil，
// 下一轮 monitor(nil, f) 在 nil channel 上永久阻塞（只剩退出信号），
// 配置热更静默死亡到进程结束且没有任何告警。
// 这里用空 prefix 让 watch 必然失败，稳定复现失败分支。
func TestReconnect_FailureKeepsChannel(t *testing.T) {
	d := NewEtcdSync()
	d.prefix = ""   // watch 拒绝空 prefix，保证走重连失败分支
	close(d.exitCh) // 首次失败后即退出重试循环，避免测试阻塞

	old := make(clientv3.WatchChan)
	ch, ok := d.reconnect(old, func(string, []byte) {})
	if ok {
		t.Fatalf("空 prefix 下 reconnect 不应成功")
	}
	if ch == nil {
		t.Fatalf("重连失败时必须保留原 channel：置空会让下一轮 monitor 永久阻塞")
	}
}
