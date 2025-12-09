package entity

import (
	"framework/library/mlog"
	"framework/library/util"
	"framework/packet"
	"path/filepath"
	"time"

	"github.com/nats-io/nats.go"
)

var (
	disconErr = nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
		mlog.Error(0, "NATS disconnect error:%v", err)
	})
	reconErr = nats.ReconnectErrHandler(func(_ *nats.Conn, err error) {
		mlog.Error(0, "NATS Reconnect error:%v", err)
	})
	close = nats.ClosedHandler(func(_ *nats.Conn) {
		mlog.Error(0, "NATS Close")
	})
	waitRecon = nats.ReconnectWait(5)
)

type NatsBus struct {
	client *nats.Conn
	prefix string
}

func NewNatsBus(prefix string, endpoints string) (ret *NatsBus, err error) {
	ret = &NatsBus{prefix: prefix}
	err = util.Retry(3, time.Second, func() error {
		ret.client, err = nats.Connect(endpoints, close, waitRecon, reconErr, disconErr)
		return err
	})
	return
}

func (d *NatsBus) Subscribe(topic string, f func(*packet.Message)) error {
	_, err := d.client.Subscribe(filepath.Join(d.prefix, topic), func(msg *nats.Msg) {
		f(&packet.Message{Reply: msg.Reply, Body: msg.Data})
	})
	return err
}

func (d *NatsBus) Send(topic string, body []byte) error {
	return d.client.Publish(filepath.Join(d.prefix, topic), body)
}

func (d *NatsBus) Request(topic string, body []byte, cb func([]byte) error) error {
	resp, err := d.client.Request(filepath.Join(d.prefix, topic), body, 3*time.Second)
	if err != nil {
		return err
	}
	return cb(resp.Data)
}

func (d *NatsBus) Response(topic string, body []byte) error {
	return d.client.Publish(topic, body)
}

func (d *NatsBus) Close() {
	d.client.Close()
}
