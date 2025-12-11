package entity

import (
	"framework/define"
	"framework/internal/global"
	"framework/library/mlog"
	"framework/library/uerror"
	"framework/packet"
	"net"
)

type Client struct {
	define.IFrame
	conn      net.Conn
	socketId  uint32      // socketId
	readBytes []byte      // 读缓存
	write     chan []byte // 写队列(控制发送速率)
	exit      chan struct{}
	list      chan<- uint32
}

func NewClient(list chan<- uint32) *Client {
	return &Client{
		socketId:  global.GenerateSocketId(),
		readBytes: make([]byte, 1024*10),
		write:     make(chan []byte, 100),
		list:      list,
		exit:      make(chan struct{}),
	}
}

func (d *Client) Init(c net.Conn, f define.IFrame) {
	d.conn = c
	d.IFrame = f
	go d.writeLoop()
}

func (d *Client) Close() {
	close(d.exit)
	d.conn.Close()
}

func (d *Client) Stop() {
	d.list <- d.socketId
}

func (d *Client) GetId() uint32 {
	return d.socketId
}

// 写循环
func (d *Client) writeLoop() {
	for {
		select {
		case buf := <-d.write:
			if _, err := d.conn.Write(buf); err != nil {
				mlog.Error(0, "socket-client write error: %v", err)
				d.Stop()
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
	case d.write <- buf:
		return nil
	case <-d.exit:
		return uerror.New(-1, "网络已经关闭")
	default:
		d.Stop()
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
				mlog.Error(0, "socket-client read error: %v", err)
				d.Stop()
				return
			}

			// 解包
			pac := d.Decode(d.readBytes[:size])
			pac.Head.SocketId = d.socketId

			// 处理请求
			if err := f(pac); err != nil {
				mlog.Error(0, "包处理错误 packet:%v, error:%v", pac, err)
			}
		}
	}
}
