package socket

import (
	"net"

	"github.com/hechh/framework"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/uerror"
)

type SocketClient struct {
	framework.IFrame
	limit  int
	conn   net.Conn
	rbytes []byte
}

func NewSocketClient(c net.Conn, f framework.IFrame, limit int) *SocketClient {
	return &SocketClient{
		IFrame: f,
		limit:  limit,
		conn:   c,
		rbytes: make([]byte, limit),
	}
}

func (d *SocketClient) Close() error {
	return d.conn.Close()
}

func (d *SocketClient) Write(pack *packet.Packet) (int, error) {
	// 获取数据帧长度
	buf := d.Encode(pack)
	if len(buf) >= d.limit {
		return 0, uerror.Err(-1, "超过最大包长限制: %d", d.limit)
	}
	// 发送数据包
	return d.conn.Write(buf)
}

func (d *SocketClient) Read() (*packet.Packet, error) {
	n, err := d.conn.Read(d.rbytes)
	if err != nil {
		return nil, err
	}
	return d.Decode(d.rbytes[:n])
}
