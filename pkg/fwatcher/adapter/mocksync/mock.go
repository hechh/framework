package mocksync

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/fwatcher/adapter/etcdsync"
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

func NewMonitor() *EmbedSync {
	return &EmbedSync{EtcdSync: etcdsync.NewEtcdSync()}
}

// Init 启动嵌入式 etcd 服务并委托 Configure.Init 完成客户端初始化。
func (m *EmbedSync) Init(cfg *fwatcher.Config) error {
	dir, err := os.MkdirTemp("", "mock-etcd-configure-")
	if err != nil {
		return fmt.Errorf("mock configure: create temp dir: %w", err)
	}
	m.dir = dir

	clientPort, peerPort, err := m.allocatePorts()
	if err != nil {
		os.RemoveAll(dir)
		return err
	}

	embedCfg := embed.NewConfig()
	embedCfg.Dir = dir
	embedCfg.LogLevel = "fatal"

	lcURL, _ := url.Parse(fmt.Sprintf("http://%s", clientPort))
	lpURL, _ := url.Parse(fmt.Sprintf("http://%s", peerPort))
	embedCfg.ListenClientUrls = []url.URL{*lcURL}
	embedCfg.ListenPeerUrls = []url.URL{*lpURL}
	embedCfg.AdvertiseClientUrls = []url.URL{*lcURL}
	embedCfg.AdvertisePeerUrls = []url.URL{*lpURL}
	embedCfg.InitialCluster = fmt.Sprintf("default=http://%s", peerPort)

	m.server, err = embed.StartEtcd(embedCfg)
	if err != nil {
		os.RemoveAll(dir)
		return fmt.Errorf("mock configure: start embedded server: %w", err)
	}

	select {
	case <-m.server.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		m.server.Close()
		os.RemoveAll(dir)
		return fmt.Errorf("mock configure: server start timeout")
	}

	// 将嵌入式 etcd 的实际地址写入 config，再委托 EtcdSync.Init 完成客户端初始化
	actualEndpoint := fmt.Sprintf("http://%s", clientPort)
	cfg.Etcd.Endpoints = []string{actualEndpoint}
	return m.EtcdSync.Init(cfg)
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
