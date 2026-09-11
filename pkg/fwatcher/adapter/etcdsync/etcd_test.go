package etcdsync

import (
	"testing"

	"github.com/hechh/framework/pkg/fwatcher"
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
