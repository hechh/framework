package entity

import (
	"net"
	"time"

	"github.com/hechh/library/uerror"

	"github.com/gorilla/websocket"
)

type WebsocketWrapper struct {
	*websocket.Conn
}

func NewWebsocketWrapper(c *websocket.Conn) net.Conn {
	return &WebsocketWrapper{Conn: c}
}

func (d *WebsocketWrapper) SetDeadline(t time.Time) error {
	return uerror.New(-1, "websocket不支持SetDeadline功能")
}

func (d *WebsocketWrapper) Read(val []byte) (int, error) {
	_, data, err := d.Conn.ReadMessage()
	copy(val, data)
	return len(data), err
}

func (d *WebsocketWrapper) Write(val []byte) (int, error) {
	err := d.Conn.WriteMessage(websocket.BinaryMessage, val)
	return len(val), err
}
