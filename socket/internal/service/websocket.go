package service

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/socket/internal/entity"
	"github.com/hechh/library/async"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/uerror"
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
	framework.IFrame
	handler func(*packet.Packet) error
	mutex   sync.RWMutex
	sockets map[uint32]framework.ISocket
	exit    chan struct{}
}

func NewWebsocket() *Websocket {
	return &Websocket{
		sockets: make(map[uint32]framework.ISocket),
		exit:    make(chan struct{}),
	}
}

func (d *Websocket) accept(conn *websocket.Conn) {
	// 创建连接
	cli := entity.NewClient()
	cli.Init(entity.NewWebsocketWrapper(conn), d)
	d.mutex.Lock()
	d.sockets[cli.GetId()] = cli
	d.mutex.Unlock()

	// 启动连接
	cli.Read(d.handler)
}

// 启动服务
func (d *Websocket) Init(f framework.IFrame, h func(*packet.Packet) error, addr string) error {
	d.IFrame = f
	d.handler = h
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil || conn == nil {
			mlog.Errorf("WebSocket连接失败: %v", err)
			return
		}
		d.accept(conn)
	})

	go func() {
		defer close(d.exit)
		mlog.Infof("启动Websocket服务启动%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}()

	async.Go(d.remove)
	return nil
}

// 关闭服务
func (d *Websocket) Close() {
	close(d.exit)
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	for _, cli := range d.sockets {
		cli.Close()
	}
}

func (d *Websocket) Get(id uint32) framework.ISocket {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if item, ok := d.sockets[id]; ok {
		return item
	}
	return nil
}

func (d *Websocket) Stop(id uint32) {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	if item, ok := d.sockets[id]; ok {
		item.Close()
		mlog.Tracef("[socket] 关闭链接%d", id)
	}
}

func (d *Websocket) remove() {
	tt := time.NewTicker(time.Duration(framework.HeartTimeExpire) * time.Second / 2)
	defer tt.Stop()
	for {
		select {
		case now := <-tt.C:
			ids := []uint32{}
			d.mutex.RLock()
			for _, cli := range d.sockets {
				if cli.IsExpire(now.Unix(), framework.HeartTimeExpire) {
					ids = append(ids, cli.GetId())
				}
			}
			d.mutex.RUnlock()

			if len(ids) > 0 {
				d.mutex.Lock()
				for _, id := range ids {
					delete(d.sockets, id)
				}
				d.mutex.Unlock()
			}
		case <-d.exit:
			return
		}
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
