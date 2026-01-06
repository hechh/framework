package service

import (
	"fmt"
	"framework/core/define"
	"framework/core/socket/internal/entity"
	"framework/library/async"
	"framework/library/mlog"
	"framework/library/uerror"
	"framework/packet"

	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	WriteBufferPool: &sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Websocket struct {
	frame   define.IFrame
	handler func(*packet.Packet) error
	mutex   sync.RWMutex
	sockets map[uint32]define.ISocket
	close   chan uint32
	exit    chan struct{}
}

func NewWebsocket(ff define.IFrame, pp func(*packet.Packet) error) *Websocket {
	return &Websocket{
		sockets: make(map[uint32]define.ISocket),
		frame:   ff,
		handler: pp,
		close:   make(chan uint32, 100),
		exit:    make(chan struct{}),
	}
}

// 启动服务
func (d *Websocket) Init(ip string, port int) error {
	if len(ip) <= 0 {
		ip = "127.0.0.1"
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil || conn == nil {
			mlog.Error(0, "WebSocket连接失败: %v", err)
			return
		}
		d.accept(conn)
	})

	go func() {
		addr := fmt.Sprintf("%s:%d", ip, port)
		mlog.Error(0, "启动Websocket服务启动%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}()

	async.Go(d.refresh)
	return nil
}

func (d *Websocket) accept(conn *websocket.Conn) {
	// 创建连接
	cli := entity.NewClient(d.close)
	cli.Init(entity.NewWebsocketWrapper(conn), d.frame)
	d.mutex.Lock()
	d.sockets[cli.GetId()] = cli
	d.mutex.Unlock()

	// 启动连接
	cli.Read(d.handler)
}

// 关闭服务
func (d *Websocket) Close() {
	close(d.exit)
	d.mutex.Lock()
	defer d.mutex.Unlock()
	for _, cli := range d.sockets {
		cli.Close()
	}
}

func (d *Websocket) refresh() {
	for {
		select {
		case socketId := <-d.close:
			d.Remove(socketId)
		case <-d.exit:
			return
		}
	}
}

func (d *Websocket) Remove(socketId uint32) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if cli, ok := d.sockets[socketId]; ok {
		cli.Close()
		delete(d.sockets, socketId)
	}
}

func (d *Websocket) Send(pac *packet.Packet) error {
	d.mutex.RLock()
	cli, ok := d.sockets[pac.Head.SocketId]
	d.mutex.RUnlock()
	if !ok {
		return uerror.New(-1, "Socket(%d:%d)不存在", pac.Head.Id, pac.Head.SocketId)
	}
	return cli.Write(pac)
}
