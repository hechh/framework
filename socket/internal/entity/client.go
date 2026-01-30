package entity

import (
	"net"
	"sync/atomic"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/uerror"
)

type Client struct {
	framework.IFrame
	conn       net.Conn      // 链接
	readBytes  []byte        // 读缓存
	writeCh    chan []byte   // 写队列(控制发送速率)
	exit       chan struct{} // 退出
	socketId   uint32        // socketId
	status     uint32        // 状态
	updateTime int64         // 更新时间
}

func NewClient() *Client {
	return &Client{
		socketId:  framework.GenSocketId(),
		readBytes: make([]byte, 1024*10),
		writeCh:   make(chan []byte, 100),
		exit:      make(chan struct{}),
	}
}

func (d *Client) Init(c net.Conn, f framework.IFrame) {
	if atomic.CompareAndSwapUint32(&d.status, 0, 1) {
		d.conn = c
		d.IFrame = f
		go d.loop()
	}
}

func (d *Client) Close() {
	if atomic.CompareAndSwapUint32(&d.status, 1, 0) {
		close(d.exit)
		d.conn.Close()
	}
}

func (d *Client) GetId() uint32 {
	return d.socketId
}

func (d *Client) IsExpire(now int64, expire int64) bool {
	if atomic.LoadUint32(&d.status) <= 0 {
		return true
	}
	atomic.CompareAndSwapInt64(&d.updateTime, 0, now)
	return now-atomic.LoadInt64(&d.updateTime) >= expire
}

func (d *Client) loop() {
	for {
		select {
		case buf := <-d.writeCh:
			if _, err := d.conn.Write(buf); err != nil {
				mlog.Errorf("发送客户端消息错误: %v", err)
				d.Close()
				return
			}
		case <-d.exit:
			return
		}
	}
}

// 发送请求
func (d *Client) Write(pac *packet.Packet) error {
	buf := d.Encode(pac)
	select {
	case d.writeCh <- buf:
		return nil
	case <-d.exit:
		return uerror.New(-1, "网络已经关闭")
	default:
		d.Close()
		return uerror.New(-1, "socket-client 网络拥堵，请求发送失败")
	}
}

// 请求转发处理
func (d *Client) Read(f func(*packet.Packet) error) {
	for {
		select {
		case <-d.exit:
			return
		default:
			size, err := d.conn.Read(d.readBytes)
			if err != nil {
				mlog.Errorf("socket-client read error: %v", err)
				d.Close()
				return
			}

			// 解包
			pac, err := d.Decode(d.readBytes[:size])
			if err != nil {
				mlog.Errorf("Packet:%v, error:%v", pac, err)
				continue
			}
			pac.Head.SocketId = d.socketId

			// 处理请求
			if err := f(pac); err != nil {
				mlog.Errorf("包处理错误 packet:%v, error:%v", pac, err)
				continue
			}
			atomic.StoreInt64(&d.updateTime, time.Now().Unix())
		}
	}
}
