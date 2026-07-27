package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/timer"
)

// IClient 网络客户端接口
type IClient interface {
	Start()                          // 启动客户端读写循环
	Stop()                           // 关闭客户端
	GetId() uint32                   // 获取 socketId
	GetUid() uint64                  // 获取uid
	SetUid(uint64) bool              // 设置uid
	Send(*packet.Head, []byte) error // 发送消息
}

// Server WebSocket 服务端
type Server struct {
	uids     map[uint64]uint32  // socket---uid绑定
	links    map[uint32]IClient // 连接池
	exitCh   chan struct{}      // 退出
	upgrader websocket.Upgrader // HTTP → WebSocket 升级器
	server   *http.Server       // HTTP 服务器
	wg       sync.WaitGroup     // 等待
	mutex    sync.RWMutex       // 加解锁
}

// NewServer 创建 WebSocket 服务端
func NewServer() *Server {
	return &Server{
		uids:   make(map[uint64]uint32),
		links:  make(map[uint32]IClient),
		exitCh: make(chan struct{}),
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
func (d *Server) Init(t *timer.Timer, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := d.upgrader.Upgrade(w, r, nil)
		if err != nil {
			mlog.Errorf("升级websocket连接失败, error:%v", err)
			return
		}

		// 创建连接
		client := NewOptimizeClient(d, conn, httpcli.GetRealIP(r))
		d.Add(client)

		// 加入定时器中
		client.Refresh(datetime.NowUnixMilli())
		t.Register(client)

		// 阻塞等待消息处理
		client.Start()

		// 连接断开处理
		d.Unbind(client.GetId(), 0)
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
func (d *Server) Close() {
	// 关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if d.server != nil {
		d.server.Shutdown(ctx)
	}

	// 关闭所有连接
	d.mutex.Lock()
	for id, cli := range d.links {
		cli.Stop()
		mlog.Infof("删除连接. socketId=%d, uid=%d", id, cli.GetUid())
	}
	d.mutex.Unlock()
	mlog.Infof("WebSocket服务器已停止")
}

func (d *Server) Add(cli IClient) {
	d.mutex.Lock()
	d.links[cli.GetId()] = cli
	d.mutex.Unlock()
}

func (d *Server) Get(id uint32, uid uint64) IClient {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if uid > 0 {
		if sid, ok := d.uids[uid]; ok {
			id = sid
		}
	}
	if cli, ok := d.links[id]; ok {
		return cli
	}
	return nil
}

func (d *Server) Bind(socketId uint32, uid uint64) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	client, ok := d.links[socketId]
	if !ok {
		return false
	}
	if client.SetUid(uid) {
		d.uids[uid] = socketId
		return true
	}
	return false
}

func (d *Server) Unbind(id uint32, uid uint64) {
	if cli := d.Get(id, uid); cli != nil {
		d.mutex.Lock()
		delete(d.uids, cli.GetUid())
		delete(d.links, cli.GetId())
		d.mutex.Unlock()
		cli.Stop()
		mlog.Infof("删除连接. socketId=%d, uid=%d", cli.GetId(), cli.GetUid())
	}
}

func (d *Server) Send(head *packet.Head, body []byte) error {
	if cli := d.Get(head.SocketId, head.Uid); cli != nil {
		return cli.Send(head, body)
	}
	return fmt.Errorf("websocket连接不存在 uid=%d, socketId=%d, cmd=%d", head.Uid, head.SocketId, head.Cmd)
}
