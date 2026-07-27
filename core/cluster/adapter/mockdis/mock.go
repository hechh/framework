package mockdis

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/cluster/adapter/discovery"
	"go.etcd.io/etcd/server/v3/embed"
)

// MockDiscovery 基于嵌入式 etcd 的 Mock 发现服务。
// 嵌入 *etcd.Etcd 以复用 Register/Watch/Close 等逻辑，仅覆盖 Init。
type MockDiscovery struct {
	*discovery.Etcd
	server *embed.Etcd
	dir    string
	once   sync.Once
}

func New() *MockDiscovery {
	return &MockDiscovery{
		Etcd: discovery.NewEtcd(),
	}
}

// Init 启动嵌入式 etcd 服务并委托 Etcd.Init 完成客户端初始化。
func (m *MockDiscovery) Init(cfg *cluster.Config) error {
	dir, err := os.MkdirTemp("", "mock-etcd-")
	if err != nil {
		return fmt.Errorf("mock etcd: create temp dir: %w", err)
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
		return fmt.Errorf("mock etcd: start embedded server: %w", err)
	}

	select {
	case <-m.server.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		m.server.Close()
		os.RemoveAll(dir)
		return fmt.Errorf("mock etcd: server start timeout")
	}

	// 将嵌入式 etcd 的本地地址作为 endpoint，委托 Etcd.Init 完成客户端初始化
	return m.Etcd.Init(&cluster.Config{
		Etcd: &cluster.EtcdConfig{
			Prefix:    cfg.Etcd.Prefix,
			Endpoints: []string{clientPort},
			KeepAlive: cfg.Etcd.KeepAlive,
		},
	})
}

// Close 关闭服务并清理资源（幂等，可安全重复调用）。
func (m *MockDiscovery) Close() {
	m.once.Do(func() {
		m.Etcd.Close()
		if m.server != nil {
			m.server.Close()
		}
		if m.dir != "" {
			os.RemoveAll(m.dir)
		}
	})
}

// allocatePorts 预绑定两个可用端口，返回 "host:port" 格式的地址。
func (m *MockDiscovery) allocatePorts() (clientAddr, peerAddr string, err error) {
	clientLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("mock etcd: bind client port: %w", err)
	}
	clientAddr = clientLn.Addr().String()
	clientLn.Close()

	peerLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", "", fmt.Errorf("mock etcd: bind peer port: %w", err)
	}
	peerAddr = peerLn.Addr().String()
	peerLn.Close()

	return clientAddr, peerAddr, nil
}
