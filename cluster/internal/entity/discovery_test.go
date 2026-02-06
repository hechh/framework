package entity

import (
	"context"
	"testing"
	"time"

	"github.com/hechh/framework"
	"github.com/hechh/library/yaml"
	"github.com/spf13/cast"
)

func TestDis(t *testing.T) {
	cfg := &yaml.EtcdConfig{
		Topic:     "/data/test",
		Endpoints: []string{"http://localhost:12379"},
	}

	cli := &EtcdDiscovery{}
	if err := cli.Init(cfg); err != nil {
		t.Log("========>", err)
		return
	}

	go func() {
		tick := time.NewTicker(2 * time.Second)
		for {
			now := <-tick.C
			t.Log("--------------->", now.Unix())
			str := cast.ToString(now.Unix())
			if err := cli.Put("test", []byte(str)); err != nil {
				t.Log("Put error", err)
				return
			}

			err := cli.Get(func(key string, val []byte) {
				t.Log(key, "------->", string(val))
			})
			if err != nil {
				t.Log("Get error", err)
				return
			}
		}
	}()

	if err := cli.KeepAlive(); err != nil {
		t.Log("Grant: ", err)
		return
	}
}

func TestDis2(t *testing.T) {
	cfg := &yaml.EtcdConfig{
		Topic:     "/data/test",
		Endpoints: []string{"http://localhost:12379"},
	}

	cli := &EtcdDiscovery{}
	if err := cli.Init(cfg); err != nil {
		t.Log("========>", err)
		return
	}

	// 租赁
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if rsp, err := cli.client.Grant(ctx, framework.ETCD_GRANT_TTL); err == nil {
		cli.lease = rsp.ID
	} else {
		t.Log("========>", err)
		return
	}

	// 设置kv
	if err := cli.Put("test", []byte("11111")); err != nil {
		t.Log("Put error", err)
		return
	}

	/*
		if _, err := cli.client.Revoke(context.Background(), cli.lease); err != nil {
			t.Log("Get error", err)
			return
		}
	*/

	if err := cli.Get(func(key string, val []byte) {
		t.Log(key, "------->", string(val))
	}); err != nil {
		t.Log("Get error", err)
		return
	}
}
