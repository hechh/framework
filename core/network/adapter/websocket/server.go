package websocket

import (
	"context"
	"net/http"
	"time"

	"github.com/hechh/framework/core/network/internal/linkpool"

	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"

	"github.com/gorilla/websocket"
)

// Server WebSocket 服务端
type Server struct {
	*linkpool.LinkPool
	upgrader websocket.Upgrader // HTTP → WebSocket 升级器
	server   *http.Server       // HTTP 服务器
}

// NewServer 创建 WebSocket 服务端
func NewServer() *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Init 初始化并启动 WebSocket 服务（实现 IServer / IComponent 接口）
func (d *Server) Init(addr string) error {
	d.LinkPool = linkpool.NewLinkPool()
	d.LinkPool.Init() // 启动自动清理协程

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := d.upgrader.Upgrade(w, r, nil)
		if err != nil {
			mlog.Errorf("升级websocket连接失败, error:%v", err)
			return
		}

		// 创建连接
		socketId := d.GenSocketId()
		client := NewOptimizeClient(conn, socketId, httpcli.GetRealIP(r))
		d.Add(client)

		// 阻塞等待消息处理
		client.Init()

		// 连接断开处理
		d.Del(socketId)
	})
	d.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		mlog.Infof("WebSocket服务器启动: %s", addr)
		if err := d.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				mlog.Errorf("WebSocket服务器错误. addr:%s, error:%v", addr, err)
			}
		}
	}()
	return nil
}

// Close 关闭 WebSocket 服务
func (s *Server) Close() {
	// 关闭所有连接
	s.LinkPool.Close()

	// 关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s.server != nil {
		s.server.Shutdown(ctx)
	}
	mlog.Infof("WebSocket服务器已停止")
}
