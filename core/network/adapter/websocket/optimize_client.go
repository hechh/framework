package websocket

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/gc"
	"github.com/hechh/framework/pkg/mlog"
	"golang.org/x/time/rate"
)

const (
	SEND_RATE_LIMIT = 100 // 每秒最多发送次数（optimize_client 使用）
	SEND_RATE_BURST = 20  // 突发允许量（optimize_client 使用）

	// CLIENT_IDLE_TTL_MS 连接闲置超时（毫秒）：超过该时长没有收到上行消息即视为死连接并断开。
	// 取 15s 的依据：约为客户端心跳间隔（2.5s）的 6 倍，且晚于游戏侧无心跳容忍值
	// （gatesrv PLAYER_KEEP_ALIVE=12s），避免在游戏自身判定超时之前就掐断连接。
	// 注意：仅上行消息会刷新活跃时间，下行发送（Send）不续期。
	CLIENT_IDLE_TTL_MS = int64(20 * time.Second / time.Millisecond)
)

// OptimizeClient WebSocket 客户端连接（优化版：无写协程，直接写 + 限流）
type OptimizeClient struct {
	parent      *Server         // 上级指针
	sendLimiter *rate.Limiter   // 发送限流器
	conn        *websocket.Conn // WebSocket 连接
	socketId    uint32          // 连接 ID
	ip          string          // 客户端真实 IP（IPv4 字符串）
	exitCh      chan struct{}   // 退出信号
	uid         atomic.Uint64   // 绑定的用户 ID
	updateTime  atomic.Int64    // 最后活跃时间（Unix 秒）
	status      atomic.Bool     // 关闭标记
	ttlMs       int64           // 时间戳
	times       int32           // 任务执行次数
}

// NewOptimizeClient 创建客户端
func NewOptimizeClient(parent *Server, conn *websocket.Conn, ip string) *OptimizeClient {
	return &OptimizeClient{
		parent:      parent,
		sendLimiter: rate.NewLimiter(SEND_RATE_LIMIT, SEND_RATE_BURST),
		conn:        conn,
		socketId:    domain.GenSocketId(),
		ip:          ip,
		exitCh:      make(chan struct{}),
		ttlMs:       CLIENT_IDLE_TTL_MS,
		times:       1,
	}
}

// 实现定时器的 Iask 接口
func (d *OptimizeClient) IsEnable() bool {
	return atomic.LoadInt32(&d.times) > 0
}
func (d *OptimizeClient) GetTTL() int64 {
	return d.ttlMs
}
func (d *OptimizeClient) GetExpire() int64 {
	return d.updateTime.Load() + d.ttlMs
}
func (d *OptimizeClient) Refresh(now int64) {
	d.updateTime.Store(now)
}
func (d *OptimizeClient) Call() {
	atomic.AddInt32(&d.times, -1)
	gc.Destroy(func() {
		// 闲置超时：主动断开连接。conn.Close() 会让阻塞中的 ReadMessage 立即返回错误，
		// readLoop 退出后由 server.go 的 Unbind 完成 links/uids 清理（status 的 CAS 保证重复 Stop 幂等）
		d.Stop()
		mlog.Warnf("WebSocket连接闲置超时断开. socketId=%d, uid=%d", d.socketId, d.uid.Load())
	})
}

// Init 启动读循环（实现 IClient 接口）
func (d *OptimizeClient) Start() {
	if d.status.CompareAndSwap(false, true) {
		d.readLoop()
	}
}

// Close 关闭客户端
func (d *OptimizeClient) Stop() {
	if d.status.CompareAndSwap(true, false) {
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

func (d *OptimizeClient) Send(head *packet.Head, body []byte) error {
	// status=true 表示连接已启动(活跃)，status=false 表示已停止(关闭)
	if !d.status.Load() {
		return fmt.Errorf("会话已关闭")
	}

	pack := &packet.Packet{Head: head, Body: body}
	data, err := domain.EncodeFrame(pack)
	if err != nil {
		mlog.Errorf("编码消息失败 error=%v, packet=%v", err, pack)
		return err
	}
	mlog.Tracef("编码消息成功 traceId=%d, uid=%d, createtime=%d, cmd=%d, seq=%d, version=%d",
		pack.Head.TraceId, pack.Head.Uid, pack.Head.CreateTime, pack.Head.Cmd, pack.Head.Seq, pack.Head.Version)

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
			d.Stop()
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
				d.Stop()
				return
			}
			if messageType != websocket.BinaryMessage {
				continue
			}
			d.updateTime.Store(datetime.NowUnixMilli())

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

	mlog.Tracef("解码消息成功 traceId=%d, uid=%d, createtime=%d, cmd=%d, seq=%d, version=%d",
		pack.Head.TraceId, pack.Head.Uid, pack.Head.CreateTime, pack.Head.Cmd, pack.Head.Seq, pack.Head.Version)
	return domain.PacketHandler(pack)
}
