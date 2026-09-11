package discovery

import (
	"testing"

	"github.com/hechh/framework/core/cluster"
)

// TestEtcdClose_Idempotent 验证重复 Close 不 panic。
//
// 初始化失败路径与 Cluster.Close() 都会调用 Close：修复前第二次执行 close(exitCh)
// 会 panic（close of closed channel），导致服务优雅关闭失败。
// 依赖本机 etcd（127.0.0.1:12379），不可用时跳过。
func TestEtcdClose_Idempotent(t *testing.T) {
	e := NewEtcd()
	cfg := &cluster.Config{Etcd: &cluster.EtcdConfig{
		Prefix:    "test/close-idempotent",
		Endpoints: []string{"127.0.0.1:12379"},
		KeepAlive: 12,
	}}
	if err := e.Init(cfg); err != nil {
		t.Skipf("etcd 不可用，跳过该集成用例: %v", err)
	}

	e.Close()
	e.Close() // 修复前此处 panic
}
