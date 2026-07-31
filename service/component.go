package service

import (
	"fmt"

	"github.com/hechh/framework/config"
	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/core/router"
	"github.com/hechh/framework/global"
	"github.com/hechh/library/base/templ"
	"github.com/hechh/library/dbpool"
	"github.com/hechh/library/fwatcher"
	"github.com/hechh/library/gc"
	"github.com/hechh/library/httpcli"
	"github.com/hechh/library/mlog"
	"github.com/hechh/library/msgqueue"
	"github.com/hechh/library/pprof"
	"github.com/hechh/library/redispool"
	"github.com/hechh/library/timer"
)

type Wrapper struct {
	app  *Service
	comp IComponent
}

func (d *Wrapper) Init(...msgqueue.Option) error { return d.comp.Init(d.app) }
func (d *Wrapper) Close()                        { d.comp.Close(d.app) }

// ==================== Config ====================
type Config struct {
	FileName string
	NodeType uint32
	NodeId   uint32
}

func (d *Config) Init(a *Service) error {
	// 解析配置
	cfg := a.GetConfig()
	if err := cfg.Init(d.FileName, d.NodeType, d.NodeId, global.CmdConvertor); err != nil {
		mlog.Errorf("[config] 模块初始化失败，error=%v", err)
		return err
	}
	config.SetObject(cfg)
	mlog.Infof("[config] 模块初始化成功")
	return nil
}

func (d *Config) Close(a *Service) {
	mlog.Infof("[config] 模块关闭成功")
}

// ==================== DbPool ====================

type DbPool struct{}

func (d *DbPool) Init(a *Service) error {
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

func (d *DbPool) Close(a *Service) {
	a.GetDbPool().Close()
	mlog.Infof("[dbpool] 模块关闭成功")
}

// ==================== Logger ====================

type Logger struct{}

func (d *Logger) Init(a *Service) error {
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

func (d *Logger) Close(a *Service) {
	a.GetLogger().Close()
	mlog.Infof("[logger] 日志模块关闭成功")
}

// ==================== Gc ====================

type Gc struct{}

func (d *Gc) Init(a *Service) error {
	obj := a.GetGc()
	if err := obj.Init(); err != nil {
		mlog.Errorf("[gc] GC模块初始化失败，error=%v", err)
		return err
	}

	gc.SetObject(obj)
	mlog.Infof("[gc] GC模块初始化成功")
	return nil
}

func (d *Gc) Close(a *Service) {
	a.GetGc().Close()
	mlog.Infof("[gc] GC模块关闭成功")
}

// ==================== HttpCli ====================

type HttpCli struct{}

func (d *HttpCli) Init(a *Service) error {
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

func (d *HttpCli) Close(a *Service) {
	a.GetHttpCli().Close()
	mlog.Infof("[httpcli] HTTP客户端关闭成功")
}

// ==================== Timer ====================

type Timer struct{}

func (d *Timer) Init(a *Service) error {
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

func (d *Timer) Close(a *Service) {
	a.GetTimer().Close()
	mlog.Infof("[timer] 定时器模块关闭成功")
}

// ==================== Pprof ====================

type Pprof struct{}

func (d *Pprof) Init(a *Service) error {
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

func (d *Pprof) Close(a *Service) {
	a.GetPprof().Close()
	mlog.Infof("[pprof] pprof模块关闭成功")
}

// ==================== RedisPool ====================

type RedisPool struct{}

func (d *RedisPool) Init(a *Service) error {
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

func (d *RedisPool) Close(a *Service) {
	a.GetRedisPool().Close()
	mlog.Infof("[redispool] Redis连接池关闭成功")
}

// ==================== Fwatcher ====================

type Fwatcher struct{}

func (d *Fwatcher) Init(a *Service) error {
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

func (d *Fwatcher) Close(a *Service) {
	a.GetFwatcher().Close()
	mlog.Infof("[fwatcher] 文件监听模块关闭成功")
}

// ==================== Network ====================

type Network struct{}

func (d *Network) Init(a *Service) error {
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

func (d *Network) Close(a *Service) {
	a.GetNetwork().Close()
	mlog.Infof("[network] 网络模块关闭成功")
}

// ==================== Router ====================

type Router struct{}

func (d *Router) Init(a *Service) error {
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

func (d *Router) Close(a *Service) {
	a.GetRouter().Close()
	mlog.Infof("[router] 路由模块关闭成功")
}

// ==================== Cluster ====================

type Cluster struct{}

func (d *Cluster) Init(a *Service) error {
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

func (d *Cluster) Close(a *Service) {
	a.GetCluster().Close()
	mlog.Infof("[cluster] 集群模块关闭成功")
}

// ==================== MsgBus ====================

type MsgBus struct{}

func (d *MsgBus) Init(a *Service) error {
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

func (d *MsgBus) Close(a *Service) {
	a.GetMsgBus().Close()
	mlog.Infof("[msgbus] 消息总线模块关闭成功")
}
