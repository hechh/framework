package network_mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	ws "github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/adapter/websocket"
	"github.com/hechh/framework/core/network/internal/linkpool"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
)

// Server 测试用 Mock WebSocket 服务端（基于 httptest）
type Server struct {
	*linkpool.LinkPool
	upgrader ws.Upgrader      // HTTP → WebSocket 升级器
	server   *httptest.Server // httptest 服务器
}

// NewServer 创建 Mock 服务端
func NewServer() *Server {
	return &Server{
		upgrader: ws.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Init 初始化并启动 Mock WebSocket 服务（实现 IServer / IComponent 接口）
func (d *Server) Init(cfg *packet.Config) error {
	nodeCfg := cfg.Node
	addr := fmt.Sprintf("%s:%d", nodeCfg.Ip, nodeCfg.Port)

	d.LinkPool = linkpool.NewLinkPool()
	// 启动自动清理协程
	d.LinkPool.Init()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := d.upgrader.Upgrade(w, r, nil)
		if err != nil {
			mlog.Errorf("升级websocket连接失败, error:%v", err)
			return
		}

		// 创建连接
		socketId := d.GenSocketId()
		client := websocket.NewOptimizeClient(conn, socketId, "127.0.0.1")
		d.Add(client)

		// 阻塞等待消息处理
		client.Init()

		// 连接断开处理
		d.Del(socketId)
	})
	d.server = httptest.NewServer(mux)
	mlog.Infof("[Mock] 模拟服务器启动: %s (httptest + WebSocket)", addr)
	return nil
}

// Close 关闭 Mock 服务端
func (s *Server) Close() {
	// 关闭所有连接
	s.LinkPool.Close()

	// 关闭服务
	s.server.Close()
	mlog.Infof("[Mock] WebSocket服务器已停止")
}
