package websocket

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hechh/framework/core/network/internal/domain"
	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/library/queue"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/packet"
	"github.com/hechh/framework/pkg/gc"
	"github.com/hechh/framework/pkg/mlog"
)

const (
	// SEND_QUEUE_MAX_BYTES 单连接待发队列软上限（字节）。
	// 队列把网络写从调用方协程（NATS 派发 / actor）解耦；达到上限即拒绝入队并显式报错，
	// 避免慢客户端让待发帧无界堆积。按「每连接预算 × 连接数」估算：128KB × 5000 连接 ≈ 640MB 上界，
	// 且慢连接会被写截止时间断开（见 WRITE_TIMEOUT_MS），常态远低于该上界。
	SEND_QUEUE_MAX_BYTES = 128 << 10

	// WRITE_TIMEOUT_MS 单帧写截止时间（毫秒）。
	// 客户端读不动时 TCP 窗口填满，写会长时间阻塞；超时判定该连接不可用并断开
	// （gorilla 语义：写超时后连接状态已损坏，只能关闭）。
	WRITE_TIMEOUT_MS = 5000

	// CLIENT_IDLE_TTL_MS 连接闲置超时（毫秒）：超过该时长没有收到上行消息即视为死连接并断开。
	// 取 15s 的依据：约为客户端心跳间隔（2.5s）的 6 倍，且晚于游戏侧无心跳容忍值
	// （gatesrv PLAYER_KEEP_ALIVE=12s），避免在游戏自身判定超时之前就掐断连接。
	// 注意：仅上行消息会刷新活跃时间，下行发送（Send）不续期。
	CLIENT_IDLE_TTL_MS = int64(20 * time.Second / time.Millisecond)
)

// OptimizeClient WebSocket 客户端连接（每连接一个读协程 + 一个写协程）。
// 读协程即处理 HTTP 升级的协程（见 server.go 的 Start 调用点），写协程由 Start 启动：
// 下行帧一律经本连接待发队列交由写协程串行写出，满足 gorilla「同一连接只有一个协程写」的约束，
// 同时调用方（NATS 派发 / actor）不再被网络写阻塞。
type OptimizeClient struct {
	parent     *Server              // 上级指针
	sendQ      *queue.Queue[[]byte] // 待发帧队列（Send 入队，本连接写协程出队）
	sendWake   chan struct{}        // 写协程唤醒信号（缓冲 1，避免通知无界积压）
	conn       *websocket.Conn      // WebSocket 连接
	socketId   uint32               // 连接 ID
	ip         string               // 客户端真实 IP（IPv4 字符串）
	exitCh     chan struct{}        // 退出信号
	uid        atomic.Uint64        // 绑定的用户 ID
	updateTime atomic.Int64         // 最后活跃时间（Unix 毫秒）
	sendBytes  atomic.Int64         // 队列中待发字节数（软上限控制）
	dropped    atomic.Uint32        // 因队列满被丢弃的帧数（限频告警用）
	status     atomic.Bool          // 关闭标记
	ttlMs      int64                // 闲置超时时长（毫秒）
	times      int32                // 任务执行次数
}

// NewOptimizeClient 创建客户端（每连接自带一条待发队列，写协程在 Start 时启动）
func NewOptimizeClient(parent *Server, conn *websocket.Conn, ip string) *OptimizeClient {
	return &OptimizeClient{
		parent:   parent,
		sendQ:    queue.NewQueue[[]byte](),
		sendWake: make(chan struct{}, 1),
		conn:     conn,
		socketId: domain.GenSocketId(),
		ip:       ip,
		exitCh:   make(chan struct{}),
		ttlMs:    CLIENT_IDLE_TTL_MS,
		times:    1,
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

// Start 启动读写循环（实现 IClient 接口）：先起本连接的写协程，再在当前协程执行读循环。
// 当前协程即处理 HTTP 升级的协程，因此每连接共 2 个协程：1 读 + 1 写。
func (d *OptimizeClient) Start() {
	if d.status.CompareAndSwap(false, true) {
		safe.SafeGo(mlog.Fatalf, d.writeLoop)
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

// Send 发送消息：编码后入本连接待发队列，由本连接写协程串行写出。
// 调用方（NATS 派发协程、actor 任务协程、连接读协程）不再直接写 socket：
// 既不会与其他协程并发写同一连接（gorilla 要求应用自行串行化），也不会被慢客户端阻塞在网络写上。
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
	mlog.Tracef("编码消息成功 traceId=%d, uid=%d, createtime=%d, cmd=%d, seq=%d, version=%d", pack.Head.TraceId, pack.Head.Uid, pack.Head.CreateTime, pack.Head.Cmd, pack.Head.Seq, pack.Head.Version)

	// 队列达到软上限 = 本连接写不过来（慢客户端或突发）：丢弃并显式报错，避免无界堆积
	if !d.push(data) {
		if n := d.dropped.Add(1); n == 1 || n%1000 == 0 {
			mlog.Warnf("发送队列已满，丢弃消息. socketId=%d, uid=%d, cmd=%d, dropped=%d",
				d.socketId, d.uid.Load(), head.Cmd, n)
		}
		return fmt.Errorf("发送队列已满: socketId=%d, cmd=%d", d.socketId, head.Cmd)
	}
	return nil
}

// push 帧入队并唤醒写协程；超过软上限返回 false（由调用方显式失败）
func (d *OptimizeClient) push(frame []byte) bool {
	if d.sendBytes.Load() >= SEND_QUEUE_MAX_BYTES {
		return false
	}
	// 计数与入队非原子：上限是软约束，允许略微超出，避免为此引入额外锁
	d.sendBytes.Add(int64(len(frame)))
	d.sendQ.Push(frame, d.wakeup)
	return true
}

// wakeup 唤醒写协程（已有待处理通知时不重复投递，队列会被一次性排空）
func (d *OptimizeClient) wakeup() {
	select {
	case d.sendWake <- struct{}{}:
	default:
	}
}

// writeLoop 本连接唯一的写协程：串行写出待发队列中的帧
func (d *OptimizeClient) writeLoop() {
	for {
		select {
		case <-d.exitCh:
			return
		case <-d.sendWake:
			d.drain()
		}
	}
}

// drain 排空待发队列；写失败即停止排空（连接已被 Stop，剩余帧在退出时丢弃）
func (d *OptimizeClient) drain() {
	for {
		frame, ok := d.sendQ.Pop()
		if !ok {
			return
		}
		d.sendBytes.Add(-int64(len(frame)))
		if !d.writeFrame(frame) {
			return
		}
	}
}

// writeFrame 写出单个数据帧；仅由本连接写协程调用（单写入口，无需加锁）
func (d *OptimizeClient) writeFrame(frame []byte) bool {
	if !d.status.Load() {
		return false // 连接已关闭：丢弃 Stop 之后队列中的残留帧
	}

	// 每帧设写截止时间：客户端读不动时不至于让写协程长期阻塞
	d.conn.SetWriteDeadline(time.Now().Add(WRITE_TIMEOUT_MS * time.Millisecond))
	if err := d.conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
		mlog.Tracef("向客户端发送消息失败 error=%v, bodySize=%d", err, len(frame))
		// 写失败（含写超时）：连接状态已损坏，主动断开，结束本连接读写协程
		d.Stop()
		return false
	}
	return true
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
		// 帧头 uid 由客户端提供、不可信：一律以本连接绑定的 uid 为准（未登录为 0），
		// 否则客户端可伪造 uid 冒充其他在线玩家，让受害者 actor 执行任意 CMD。
		// 绑定发生在登录流程中 token 校验通过之后（network.Bind → SetUid）。
		pack.Head.Uid = d.GetUid()
	}

	mlog.Tracef("解码消息成功 traceId=%d, uid=%d, createtime=%d, cmd=%d, seq=%d, version=%d",
		pack.Head.TraceId, pack.Head.Uid, pack.Head.CreateTime, pack.Head.Cmd, pack.Head.Seq, pack.Head.Version)
	return domain.PacketHandler(pack)
}
