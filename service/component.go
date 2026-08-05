package service

import (
	"fmt"

	"github.com/hechh/framework/config"
	"github.com/hechh/framework/core/cluster"
	"github.com/hechh/framework/core/msgbus"
	"github.com/hechh/framework/core/network"
	"github.com/hechh/framework/core/router"
	"github.com/hechh/framework/global"
	"github.com/hechh/framework/packet"
	"github.com/hechh/library/base/enum"
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

// ==================== Global ====================
type Global struct {
	Cmd  enum.IConvertor
	Node enum.IConvertor
}

func (d *Global) Init(a *Service) error {
	global.CmdConvertor = d.Cmd
	global.NodeConvertor = d.Node
	mlog.Infof("[global] 初始化成功")
	return nil
}

func (d *Global) Close(a *Service) {
	mlog.Infof("[global] 关闭成功")
}

// ==================== Config ====================
type Config struct {
	File string
	Type enum.IEnum
	Id   int
}

func (d *Config) Init(a *Service) error {
	// 解析配置
	cfg := a.GetConfig()
	if err := cfg.Init(d.File, uint32(d.Type.Number()), uint32(d.Id), global.NodeConvertor); err != nil {
		mlog.Errorf("[config] 初始化失败，error=%v", err)
		return err
	}
	config.SetObject(cfg)
	global.Self = cfg.GetSelfNode()
	global.GatewayNodeType = cfg.GetGateway().Type
	mlog.Infof("[config] 初始化成功")
	return nil
}

func (d *Config) Close(a *Service) {
	mlog.Infof("[config] 关闭成功")
}

// ==================== DbPool ====================

type DbPool struct{}

func (d *DbPool) Init(a *Service) error {
	cfg := a.GetConfig()
	obj := a.GetDbPool()
	if err := obj.Init(cfg.Mysql); err != nil {
		mlog.Errorf("[dbpool] 初始化失败，error=%v", err)
		return err
	}

	dbpool.SetObject(obj)
	mlog.Infof("[dbpool] 初始化成功")
	return nil
}

func (d *DbPool) Close(a *Service) {
	a.GetDbPool().Close()
	mlog.Infof("[dbpool] 关闭成功")
}

// ==================== Logger ====================

type Logger struct{}

func (d *Logger) Init(a *Service) error {
	cfg := a.GetConfig().GetSelfConfig()
	obj := a.GetLogger()
	if err := obj.Init(cfg.Logger); err != nil {
		mlog.Errorf("[logger] 日志初始化失败，error=%v", err)
		return err
	}

	mlog.SetObject(obj)
	mlog.Infof("[logger] 日志初始化成功")
	return nil
}

func (d *Logger) Close(a *Service) {
	a.GetLogger().Close()
	mlog.Infof("[logger] 日志关闭成功")
}

// ==================== Gc ====================

type Gc struct{}

func (d *Gc) Init(a *Service) error {
	obj := a.GetGc()
	if err := obj.Init(); err != nil {
		mlog.Errorf("[gc] GC初始化失败，error=%v", err)
		return err
	}

	gc.SetObject(obj)
	mlog.Infof("[gc] GC初始化成功")
	return nil
}

func (d *Gc) Close(a *Service) {
	a.GetGc().Close()
	mlog.Infof("[gc] GC关闭成功")
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
		mlog.Errorf("[timer] 定时器初始化失败，error=%v", err)
		return err
	}

	timer.SetObject(obj)
	mlog.Infof("[timer] 定时器初始化成功")
	return nil
}

func (d *Timer) Close(a *Service) {
	a.GetTimer().Close()
	mlog.Infof("[timer] 定时器关闭成功")
}

// ==================== Pprof ====================

type Pprof struct{}

func (d *Pprof) Init(a *Service) error {
	cfg := a.GetConfig()
	obj := a.GetPprof()
	port := cfg.GetSelfNode().Port + 10000
	if err := obj.Init(int(port)); err != nil {
		mlog.Errorf("[pprof] pprof初始化失败，error=%v", err)
		return err
	}

	pprof.SetObject(obj)
	mlog.Infof("[pprof] pprof初始化成功")
	return nil
}

func (d *Pprof) Close(a *Service) {
	a.GetPprof().Close()
	mlog.Infof("[pprof] pprof关闭成功")
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
		mlog.Errorf("[fwatcher] 文件监听初始化失败，error=%v", err)
		return err
	}

	fwatcher.SetObject(obj)
	mlog.Infof("[fwatcher] 文件监听初始化成功")
	return nil
}

func (d *Fwatcher) Close(a *Service) {
	a.GetFwatcher().Close()
	mlog.Infof("[fwatcher] 文件监听关闭成功")
}

// ==================== Network ====================

type Network struct {
	Decoder func([]byte) (*packet.Packet, error)
	Encoder func(*packet.Packet) ([]byte, error)
	Handler func(*packet.Packet) error
}

func (d *Network) Init(a *Service) error {
	cfg := a.GetConfig()
	obj := a.GetNetwork()

	if d.Decoder != nil {
		network.SetDecodeFunc(d.Decoder)
	}
	if d.Encoder != nil {
		network.SetEncodeFunc(d.Encoder)
	}
	network.SetPacketFunc(d.Handler)

	addr := fmt.Sprintf(":%d", cfg.GetSelfNode().Port)
	if err := obj.Init(a.GetTimer(), addr); err != nil {
		mlog.Errorf("[network] 网络初始化失败，error=%v", err)
		return err
	}

	network.SetObject(obj)
	mlog.Infof("[network] 网络初始化成功")
	return nil
}

func (d *Network) Close(a *Service) {
	a.GetNetwork().Close()
	mlog.Infof("[network] 网络关闭成功")
}

// ==================== Router ====================

type Router struct{}

func (d *Router) Init(a *Service) error {
	cfg := a.GetConfig()
	obj := a.GetRouter()
	if err := obj.Init(a.GetTimer(), cfg.GetSelfNode(), cfg.GetSupports()); err != nil {
		mlog.Errorf("[router] 路由初始化失败，error=%v", err)
		return err
	}

	router.SetObject(obj)
	mlog.Infof("[router] 路由初始化成功")
	return nil
}

func (d *Router) Close(a *Service) {
	a.GetRouter().Close()
	mlog.Infof("[router] 路由关闭成功")
}

// ==================== Cluster ====================

type Cluster struct{}

func (d *Cluster) Init(a *Service) error {
	cfg := a.GetConfig()
	obj := a.GetCluster()
	if err := obj.Init(cfg.Discovery, cfg.GetSelfNode(), cfg.GetSupports()); err != nil {
		mlog.Errorf("[cluster] 集群初始化失败，error=%v", err)
		return err
	}

	cluster.SetObject(obj)
	mlog.Infof("[cluster] 集群初始化成功")
	return nil
}

func (d *Cluster) Close(a *Service) {
	a.GetCluster().Close()
	mlog.Infof("[cluster] 集群关闭成功")
}

// ==================== MsgBus ====================

type MsgBus struct {
	Handler func(*packet.Packet)
}

func (d *MsgBus) Init(a *Service) error {
	cfg := a.GetConfig()
	selfNode := cfg.GetSelfNode()

	msgbus.SetPacketFunc(router.RouteHandler(d.Handler))

	obj := a.GetMsgBus()
	if err := obj.Init(cfg.MsgBus, selfNode.Type, selfNode.Id); err != nil {
		mlog.Errorf("[msgbus] 消息总线初始化失败，error=%v", err)
		return err
	}

	msgbus.SetObject(obj)
	mlog.Infof("[msgbus] 消息总线初始化成功")
	return nil
}

func (d *MsgBus) Close(a *Service) {
	a.GetMsgBus().Close()
	mlog.Infof("[msgbus] 消息总线关闭成功")
}
