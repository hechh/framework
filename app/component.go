package app

import (
	"fmt"

	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/core/router"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/gc"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/pprof"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/timer"
)

type IComponent interface {
	Init(*App) error
	Close(*App)
}

// ==================== DbPool ====================

type DbPool struct{}

func (d *DbPool) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetDbPool()
	if err := obj.Init(cfg.Mysql); err != nil {
		mlog.Errorf("[dbpool] 模块初始化失败，error=%v", err)
		return err
	}

	dbpool.SetObject(obj)
	mlog.Infof("[dbpool] 模块初始化成功")
	return nil
}

func (d *DbPool) Close(a *App) {
	a.GetDbPool().Close()
	mlog.Infof("[dbpool] 模块关闭成功")
}

// ==================== Logger ====================

type Logger struct{}

func (d *Logger) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetLogger()
	if err := obj.Init(cfg.Logger); err != nil {
		mlog.Errorf("[logger] 日志模块初始化失败，error=%v", err)
		return err
	}

	mlog.SetObject(obj)
	mlog.Infof("[logger] 日志模块初始化成功")
	return nil
}

func (d *Logger) Close(a *App) {
	a.GetLogger().Close()
	mlog.Infof("[logger] 日志模块关闭成功")
}

// ==================== Gc ====================

type Gc struct{}

func (d *Gc) Init(a *App) error {
	obj := a.GetGc()
	if err := obj.Init(); err != nil {
		mlog.Errorf("[gc] GC模块初始化失败，error=%v", err)
		return err
	}

	gc.SetObject(obj)
	mlog.Infof("[gc] GC模块初始化成功")
	return nil
}

func (d *Gc) Close(a *App) {
	a.GetGc().Close()
	mlog.Infof("[gc] GC模块关闭成功")
}

// ==================== HttpCli ====================

type HttpCli struct{}

func (d *HttpCli) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetHttpCli()
	if err := obj.Init(cfg.HttpCli); err != nil {
		mlog.Errorf("[httpcli] HTTP客户端初始化失败，error=%v", err)
		return err
	}

	httpcli.SetObject(obj)
	mlog.Infof("[httpcli] HTTP客户端初始化成功")
	return nil
}

func (d *HttpCli) Close(a *App) {
	a.GetHttpCli().Close()
	mlog.Infof("[httpcli] HTTP客户端关闭成功")
}

// ==================== Timer ====================

type Timer struct{}

func (d *Timer) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetTimer()
	if err := obj.Init(cfg.Timer); err != nil {
		mlog.Errorf("[timer] 定时器模块初始化失败，error=%v", err)
		return err
	}

	timer.SetObject(obj)
	mlog.Infof("[timer] 定时器模块初始化成功")
	return nil
}

func (d *Timer) Close(a *App) {
	a.GetTimer().Close()
	mlog.Infof("[timer] 定时器模块关闭成功")
}

// ==================== Pprof ====================

type Pprof struct{}

func (d *Pprof) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetPprof()
	port := cfg.Node.Port + 10000
	if err := obj.Init(port); err != nil {
		mlog.Errorf("[pprof] pprof模块初始化失败，error=%v", err)
		return err
	}

	pprof.SetObject(obj)
	mlog.Infof("[pprof] pprof模块初始化成功")
	return nil
}

func (d *Pprof) Close(a *App) {
	a.GetPprof().Close()
	mlog.Infof("[pprof] pprof模块关闭成功")
}

// ==================== RedisPool ====================

type RedisPool struct{}

func (d *RedisPool) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetRedisPool()
	if err := obj.Init(cfg.Redis); err != nil {
		mlog.Errorf("[redispool] Redis连接池初始化失败，error=%v", err)
		return err
	}

	redispool.SetObject(obj)
	mlog.Infof("[redispool] Redis连接池初始化成功")
	return nil
}

func (d *RedisPool) Close(a *App) {
	a.GetRedisPool().Close()
	mlog.Infof("[redispool] Redis连接池关闭成功")
}

// ==================== Fwatcher ====================

type Fwatcher struct{}

func (d *Fwatcher) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetFwatcher()
	if err := obj.Init(cfg.Fwatcher); err != nil {
		mlog.Errorf("[fwatcher] 文件监听模块初始化失败，error=%v", err)
		return err
	}

	fwatcher.SetObject(obj)
	mlog.Infof("[fwatcher] 文件监听模块初始化成功")
	return nil
}

func (d *Fwatcher) Close(a *App) {
	a.GetFwatcher().Close()
	mlog.Infof("[fwatcher] 文件监听模块关闭成功")
}

// ==================== Network ====================

type Network struct{}

func (d *Network) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetNetwork()
	addr := fmt.Sprintf(":%d", cfg.Node.Port)
	if err := obj.Init(a.GetTimer(), addr); err != nil {
		mlog.Errorf("[network] 网络模块初始化失败，error=%v", err)
		return err
	}

	network.SetObject(obj)
	mlog.Infof("[network] 网络模块初始化成功")
	return nil
}

func (d *Network) Close(a *App) {
	a.GetNetwork().Close()
	mlog.Infof("[network] 网络模块关闭成功")
}

// ==================== Router ====================

type Router struct{}

func (d *Router) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetRouter()
	if err := obj.Init(a.GetTimer(), cfg.Self, templ.Map2List(cfg.Types)); err != nil {
		mlog.Errorf("[router] 路由模块初始化失败，error=%v", err)
		return err
	}

	router.SetObject(obj)
	mlog.Infof("[router] 路由模块初始化成功")
	return nil
}

func (d *Router) Close(a *App) {
	a.GetRouter().Close()
	mlog.Infof("[router] 路由模块关闭成功")
}

// ==================== Cluster ====================

type Cluster struct{}

func (d *Cluster) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetCluster()
	if err := obj.Init(cfg.Cluster, cfg.Self, templ.Map2List(cfg.Types)); err != nil {
		mlog.Errorf("[cluster] 集群模块初始化失败，error=%v", err)
		return err
	}

	cluster.SetObject(obj)
	mlog.Infof("[cluster] 集群模块初始化成功")
	return nil
}

func (d *Cluster) Close(a *App) {
	a.GetCluster().Close()
	mlog.Infof("[cluster] 集群模块关闭成功")
}

// ==================== MsgBus ====================

type MsgBus struct{}

func (d *MsgBus) Init(a *App) error {
	cfg := a.GetConfig()
	obj := a.GetMsgBus()
	if err := obj.Init(cfg.MsgBus, cfg.Node.Type, cfg.Node.Id); err != nil {
		mlog.Errorf("[msgbus] 消息总线模块初始化失败，error=%v", err)
		return err
	}

	msgbus.SetObject(obj)
	mlog.Infof("[msgbus] 消息总线模块初始化成功")
	return nil
}

func (d *MsgBus) Close(a *App) {
	a.GetMsgBus().Close()
	mlog.Infof("[msgbus] 消息总线模块关闭成功")
}
