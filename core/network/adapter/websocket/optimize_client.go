package websocket

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/datetime"
	"github.com/hechh/library/mlog"
	"golang.org/x/time/rate"
)

// OptimizeClient WebSocket 客户端连接（优化版：无写协程，直接写 + 限流）
type OptimizeClient struct {
	conn        *websocket.Conn // WebSocket 连接
	socketId    uint32          // 连接 ID
	ip          string          // 客户端真实 IP（IPv4 字符串）
	exitCh      chan struct{}   // 退出信号
	uid         atomic.Uint64   // 绑定的用户 ID
	updateTime  atomic.Int64    // 最后活跃时间（Unix 秒）
	closed      atomic.Bool     // 关闭标记
	sendLimiter *rate.Limiter   // 发送限流器
}

// NewOptimizeClient 创建客户端
func NewOptimizeClient(conn *websocket.Conn, socketId uint32, ip ...string) *OptimizeClient {
	clientIP := ""
	if len(ip) > 0 {
		clientIP = ip[0]
	}
	return &OptimizeClient{
		conn:        conn,
		socketId:    socketId,
		ip:          clientIP,
		exitCh:      make(chan struct{}),
		sendLimiter: rate.NewLimiter(SEND_RATE_LIMIT, SEND_RATE_BURST),
	}
}

// Init 启动读循环（实现 IClient 接口）
func (d *OptimizeClient) Init() {
	d.updateTime.Store(datetime.NowUnix())
	d.readLoop()
}

// Close 关闭客户端
func (d *OptimizeClient) Close() {
	if d.closed.CompareAndSwap(false, true) {
		close(d.exitCh)
		d.conn.Close()
	}
}

// GetId 获取 socketId
func (d *OptimizeClient) GetId() uint32 {
	return d.socketId
}

// Bind CAS 绑定 uid（仅一次）
func (d *OptimizeClient) SetUid(uid uint64) bool {
	return d.uid.CompareAndSwap(0, uid)
}

// Unbind 解绑 uid
func (d *OptimizeClient) GetUid() uint64 {
	return d.uid.Load()
}

// GetUpdateTime 获取最后活跃时间
func (d *OptimizeClient) GetUpdateTime() int64 {
	return d.updateTime.Load()
}

func (d *OptimizeClient) Send(head *packet.Head, body []byte) error {
	if d.closed.Load() {
		return fmt.Errorf("会话已关闭")
	}

	pack := &packet.Packet{Head: head, Body: body}
	data, err := domain.EncodeFrame(pack)
	if err != nil {
		mlog.Errorf("编码消息失败 error=%v, packet=%v", err, pack)
		return err
	}
	mlog.Tracef("编码消息成功 traceId=%d, uid=%d, createtime=%d, cmd=%d, seq=%d, version=%d, ", pack.Head.TraceId, pack.Head.Uid, pack.Head.CreateTime, pack.Head.Cmd, pack.Head.Seq, pack.Head.Version)

	if err := d.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		mlog.Tracef("向客户端发送消息失败 error=%v, bodySize=%d", err, len(data))
		return fmt.Errorf("发送失败: %w", err)
	}
	return nil
}

// readLoop 阻塞读取消息
func (d *OptimizeClient) readLoop() {
	defer func() {
		if err := recover(); err != nil {
			mlog.Fatalf("PANIC: readLoop崩溃. error:%v, socketId:%d, uid:%d\nStack Trace:\n%s", err, d.socketId, d.uid.Load(), string(debug.Stack()))
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
func (d *OptimizeClient) decodePacket(data []byte) error {
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
