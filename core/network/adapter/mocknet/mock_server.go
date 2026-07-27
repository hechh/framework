package mocknet

import (
	"net/http"
	"net/http/httptest"

	ws "github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/adapter/websocket"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
)

// Server 测试用 Mock WebSocket 服务端（基于 httptest）
// 嵌入 *websocket.Server 复用连接管理（Add/Get/Bind/Unbind/Send）
type Server struct {
	*websocket.Server                  // 复用 websocket.Server 的连接管理
	upgrader          ws.Upgrader      // HTTP → WebSocket 升级器
	server            *httptest.Server // httptest 服务器
}

// NewServer 创建 Mock 服务端
func New() *Server {
	return &Server{
		Server: websocket.NewServer(),
		upgrader: ws.Upgrader{
			ReadBufferSize:  4096,
			WriteBufferSize: 4096,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Init 初始化并启动 Mock WebSocket 服务（实现 INetwork 接口）
func (d *Server) Init(t *timer.Timer, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := d.upgrader.Upgrade(w, r, nil)
		if err != nil {
			mlog.Errorf("升级websocket连接失败, error:%v", err)
			return
		}

		// 创建连接（传入 d.Server 作为 parent）
		client := websocket.NewOptimizeClient(d.Server, conn, httpcli.GetRealIP(r))
		d.Add(client)

		// 阻塞等待消息处理
		client.Start()

		// 连接断开处理
		d.Unbind(client.GetId(), 0)
	})

	d.server = httptest.NewServer(mux)
	mlog.Infof("[Mock] 模拟服务器启动: %s (httptest + WebSocket)", addr)
	return nil
}

// Close 关闭 Mock 服务端
func (d *Server) Close() {
	// 关闭 httptest 服务
	d.server.Close()

	// 关闭所有连接
	d.Server.Close()

	mlog.Infof("[Mock] WebSocket服务器已停止")
}
