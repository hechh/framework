package websocket

import (
	"fmt"
	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/packet"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/mlog"

	"github.com/gorilla/websocket"
)

const (
	WRITE_TIMEOUT     = 10 * time.Second // 单次 WebSocket 写超时
	WRITERS_CHAN_SIZE = 256              // 发送 channel 缓冲区大小
	SEND_RATE_LIMIT   = 100              // 每秒最多发送次数（optimize_client 使用）
	SEND_RATE_BURST   = 20               // 突发允许量（optimize_client 使用）
)

// Client WebSocket 客户端连接
type Client struct {
	conn       *websocket.Conn     // WebSocket 连接
	socketId   uint32              // 连接 ID
	ip         string              // 客户端真实 IP
	exitCh     chan struct{}       // 退出信号
	uid        atomic.Uint64       // 绑定的用户 ID
	writers    chan *packet.Packet // 发送 channel（替代 lock-free queue + notifyCh + ticker）
	updateTime atomic.Int64        // 最后活跃时间（Unix 秒）
	closed     atomic.Bool         // 关闭标记
}

// NewClient 创建客户端
func NewClient(conn *websocket.Conn, socketId uint32, ip ...string) *Client {
	clientIP := ""
	if len(ip) > 0 {
		clientIP = ip[0]
	}
	return &Client{
		conn:     conn,
		socketId: socketId,
		ip:       clientIP,
		exitCh:   make(chan struct{}),
		writers:  make(chan *packet.Packet, WRITERS_CHAN_SIZE),
	}
}

// Init 启动读写循环（实现 IClient 接口）
func (d *Client) Init() {
	d.updateTime.Store(datetime.NowUnix())

	go d.writeLoop()

	d.readLoop()
}

// Close 关闭客户端
func (d *Client) Close() {
	if d.closed.CompareAndSwap(false, true) {
		close(d.exitCh)
		d.conn.Close()
	}
}

// GetId 获取 socketId
func (d *Client) GetId() uint32 {
	return d.socketId
}

// Bind CAS 绑定 uid（仅一次）
func (d *Client) SetUid(uid uint64) bool {
	return d.uid.CompareAndSwap(0, uid)
}

// Unbind 解绑 uid
func (d *Client) GetUid() uint64 {
	return d.uid.Load()
}

// GetUpdateTime 获取最后活跃时间
func (d *Client) GetUpdateTime() int64 {
	return d.updateTime.Load()
}

// Send 发送消息到客户端（channel 满时提供背压，阻止 handler 继续生产）
func (d *Client) Send(head *packet.Head, body []byte) error {
	if d.closed.Load() {
		return fmt.Errorf("会话已关闭")
	}

	select {
	case d.writers <- &packet.Packet{Head: head, Body: body}:
		return nil
	case <-d.exitCh:
		return fmt.Errorf("会话已关闭")
	}
}

// readLoop 阻塞读取消息
func (d *Client) readLoop() {
	defer func() {
		if err := recover(); err != nil {
			mlog.Fatalf("PANIC: readLoop崩溃. error:%v, socketId:%d, uid:%d\nStack Trace:\n%s",
				err, d.socketId, d.uid.Load(), string(debug.Stack()))
			d.Close()
		}
	}()

	for {
		select {
		case <-d.exitCh:
			return
		default:
			messageType, data, err := d.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					mlog.Warnf("WebSocket读取错误. error:%v, socketId:%d, uid:%d", err, d.socketId, d.uid.Load())
				}
				d.Close()
				return
			}
			if messageType != websocket.BinaryMessage {
				continue
			}
			d.updateTime.Store(datetime.NowUnix())

			if err := d.decodePacket(data); err != nil {
				mlog.Errorf("解码数据帧失败. error:%v, socketId:%d, uid:%d", err, d.socketId, d.uid.Load())
			}
		}
	}
}

// decodePacket 解码并处理消息
func (d *Client) decodePacket(data []byte) error {
	pack, err := domain.DecodeFrame(data)
	if err != nil {
		return err
	}
	if pack.Head != nil {
		pack.Head.SocketId = d.socketId
		pack.Head.ClientIp = d.ip
	}
	return domain.PacketHandler(pack)
}

// writeLoop 写协程：从 channel 读取消息，逐帧写入 WebSocket
func (d *Client) writeLoop() {
	defer d.Close()

	for {
		select {
		case pack := <-d.writers:
			body, err := domain.EncodeFrame(pack)
			if err != nil {
				mlog.Errorf("编码消息失败 error=%v, packet=%v", err, pack)
				continue
			}

			d.conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT))
			if err := d.conn.WriteMessage(websocket.BinaryMessage, body); err != nil {
				mlog.Errorf("发送消息失败 error=%v, socketId=%d, cmd=%d", err, d.socketId, pack.Head.Cmd)
				return
			}
		case <-d.exitCh:
			return
		}
	}
}
