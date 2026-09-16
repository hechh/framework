package application_test

import (
	"testing"

	"github.com/hechh/framework/application"
	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/cluster/adapter/discovery"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/msgbus/adapter/nats"
	"github.com/hechh/framework/pkg/dbpool"
	"github.com/hechh/framework/pkg/dbpool/adapter/mysql"
	"github.com/hechh/framework/pkg/fwatcher"
	"github.com/hechh/framework/pkg/fwatcher/adapter/etcdsync"
	"github.com/hechh/framework/pkg/httpcli"
	"github.com/hechh/framework/pkg/mlog"
	"github.com/hechh/framework/pkg/timer"
	"github.com/hechh/framework/pkg/timer/adapter/lockfree_timer"
)

// TestComponentInit_MissingConfig 校验配置段缺失（Config 为 nil / Configs 为空）时组件返回明确错误而不是 panic：
// 服务入口把 application.Init 的错误交给 panic(err)，缺段理应在启动期暴露成可读提示。
func TestComponentInit_MissingConfig(t *testing.T) {
	cases := []struct {
		name string
		comp application.IComponent
	}{
		{"mlog", &mlog.Component{}},
		{"timer", &timer.Component{Object: timer.NewTimer(lockfree_timer.NewTimer)}},
		{"fwatcher", &fwatcher.Component{Object: fwatcher.NewFWatcher(etcdsync.NewEtcdSync)}},
		{"cluster", &cluster.Component{Object: cluster.NewCluster(discovery.NewEtcd())}},
		{"msgbus", &msgbus.Component{Object: msgbus.NewMsgBus(nats.NewNats())}},
		{"dbpool", &dbpool.Component{Object: dbpool.NewDbPool(mysql.NewClient)}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.comp.Init()
			if err == nil {
				t.Fatalf("%s: 配置缺失时应返回错误", c.name)
			}
			t.Logf("%s: %v", c.name, err)
		})
	}
}

// TestComponentInit_HttpcliDefaults 固定 httpcli 的既有约定：Config 为 nil 时走内置默认值，不视为配置错误。
func TestComponentInit_HttpcliDefaults(t *testing.T) {
	if err := (&httpcli.Component{}).Init(); err != nil {
		t.Fatalf("httpcli 应容忍 nil 配置并走默认值, got err=%v", err)
	}
}
