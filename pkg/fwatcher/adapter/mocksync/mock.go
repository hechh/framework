package mocksync

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/fwatcher/adapter/etcdsync"
	"github.com/hechh/framework/pkg/mlog"
	"go.etcd.io/etcd/server/v3/embed"
)

// EmbedSync 基于嵌入式 etcd 的 Mock 配置中心。
// 嵌入 *fwatcher.Configure 以复用 Put/Update/Delete/Watch 等逻辑，仅覆盖 Init。
type EmbedSync struct {
	*etcdsync.EtcdSync
	server *embed.Etcd
	dir    string
	once   sync.Once
}

const (
	// TEMP_DIR_PREFIX 嵌入式 etcd 数据目录前缀（位于系统临时目录下）
	TEMP_DIR_PREFIX = "mock-etcd-configure-"

	// STALE_DIR_TTL 遗留临时目录回收阈值。
	// 测试进程以 os.Exit(m.Run()) 结束，不会执行任何清理、EmbedSync.Close 永不触发，
	// 因此每个跑 mock 的测试进程都会遗留一个约 80MB 的 etcd 数据目录；
	// 长期累积会写满临时盘，随后 etcd 建不了 WAL（panic: failed to create WAL）、
	// Go 也写不了构建缓存（[build failed]）。故每次启动顺带回收超过该阈值的旧目录。
	// 取 30min 的依据：远超单个测试包的运行时长（秒级），不会误删正在使用的目录。
	STALE_DIR_TTL = 30 * time.Minute

	// START_MAX_ATTEMPTS 嵌入式 etcd 启动最大尝试次数。
	// allocatePorts 是「bind :0 探测 → Close」，与 etcd 自身 bind 之间存在空档（TOCTOU）：
	// 并行跑多个测试包时（每个包各起一个嵌入式 etcd）该端口可能被别的进程抢走，
	// 表现为 embed.StartEtcd 报 bind 失败并让测试直接 panic。失败即换一组端口重试，避免偶发失败打崩测试。
	START_MAX_ATTEMPTS = 5

	// START_READY_TIMEOUT 等待嵌入式 etcd 就绪的超时
	START_READY_TIMEOUT = 10 * time.Second
)

func NewMonitor() *EmbedSync {
	return &EmbedSync{EtcdSync: etcdsync.NewEtcdSync()}
}

// Init 启动嵌入式 etcd 服务并委托 EtcdSync.Init 完成客户端初始化。
// 端口被抢占等临时失败会换端口重试（见 START_MAX_ATTEMPTS）。
func (m *EmbedSync) Init(cfg *fwatcher.Config) error {
	// 先回收上轮测试遗留的数据目录（测试进程 os.Exit 结束，无法自行清理，见 STALE_DIR_TTL）
	if n := sweepStaleDirs(); n > 0 {
		mlog.Infof("mock configure: 回收遗留的嵌入式 etcd 数据目录 %d 个", n)
	}

	var lastErr error
	for attempt := 1; attempt <= START_MAX_ATTEMPTS; attempt++ {
		endpoint, err := m.startEmbedded()
		if err == nil {
			// 将嵌入式 etcd 的实际地址写入 config，再委托 EtcdSync.Init 完成客户端初始化
			cfg.Etcd.Endpoints = []string{endpoint}
			return m.EtcdSync.Init(cfg)
		}
		lastErr = err
		// 换端口重试属于异常情况（端口被抢占/资源紧张）：记录下来便于发现测试环境的并发压力
		mlog.Warnf("mock configure: 嵌入式 etcd 启动失败，换端口重试(%d/%d): %v", attempt, START_MAX_ATTEMPTS, err)
	}
	return fmt.Errorf("mock configure: 嵌入式 etcd 启动失败（已重试 %d 次）: %w", START_MAX_ATTEMPTS, lastErr)
}

// Close 关闭服务并清理资源（幂等，可安全重复调用）。
func (m *EmbedSync) Close() {
	m.once.Do(func() {
		m.EtcdSync.Close()
		if m.server != nil {
			m.server.Close()
		}
		if m.dir != "" {
			os.RemoveAll(m.dir)
		}
	})
}

// startEmbedded 申请临时目录与端口并启动嵌入式 etcd；任一步失败都会清理本次占用的资源再返回错误
func (m *EmbedSync) startEmbedded() (endpoint string, err error) {
	dir, err := os.MkdirTemp("", TEMP_DIR_PREFIX)
	if err != nil {
		return "", fmt.Errorf("mock configure: create temp dir: %w", err)
	}

	clientAddr, peerAddr, err := m.allocatePorts()
	if err != nil {
		os.RemoveAll(dir)
		return "", err
	}

	embedCfg := embed.NewConfig()
	embedCfg.Dir = dir
	embedCfg.LogLevel = "fatal"

	lcURL, _ := url.Parse(fmt.Sprintf("http://%s", clientAddr))
	lpURL, _ := url.Parse(fmt.Sprintf("http://%s", peerAddr))
	embedCfg.ListenClientUrls = []url.URL{*lcURL}
	embedCfg.ListenPeerUrls = []url.URL{*lpURL}
	embedCfg.AdvertiseClientUrls = []url.URL{*lcURL}
	embedCfg.AdvertisePeerUrls = []url.URL{*lpURL}
	embedCfg.InitialCluster = fmt.Sprintf("default=http://%s", peerAddr)

	server, err := embed.StartEtcd(embedCfg)
	if err != nil {
		os.RemoveAll(dir)
		return "", fmt.Errorf("mock configure: start embedded server: %w", err)
	}

	select {
	case <-server.Server.ReadyNotify():
	case <-time.After(START_READY_TIMEOUT):
		server.Close()
		os.RemoveAll(dir)
		return "", fmt.Errorf("mock configure: server start timeout")
	}

	m.server, m.dir = server, dir
	return fmt.Sprintf("http://%s", clientAddr), nil
}

// sweepStaleDirs 回收超过 STALE_DIR_TTL 未变更的嵌入式 etcd 数据目录，返回回收数量。
// 只在 Init 时顺带执行：单个测试包的 etcd 目录生命周期是秒级，不会误删在用目录。
func sweepStaleDirs() int {
	dirs, err := filepath.Glob(filepath.Join(os.TempDir(), TEMP_DIR_PREFIX+"*"))
	if err != nil {
		return 0
	}

	expireAt := time.Now().Add(-STALE_DIR_TTL)
	removed := 0
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil || info.ModTime().After(expireAt) {
			continue
		}
		if err := os.RemoveAll(dir); err == nil {
			removed++
		}
	}
	return removed
}

// allocatePorts 预绑定两个可用端口，返回 "host:port" 格式的地址。
func (m *EmbedSync) allocatePorts() (clientAddr, peerAddr string, err error) {
	clientLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("mock configure: bind client port: %w", err)
	}
	clientAddr = clientLn.Addr().String()
	clientLn.Close()

	peerLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("mock configure: bind peer port: %w", err)
	}
	peerAddr = peerLn.Addr().String()
	peerLn.Close()

	return clientAddr, peerAddr, nil
}
