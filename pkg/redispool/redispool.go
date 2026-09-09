package redispool

import (
	"github.com/hechh/framework/library/consistent"
)

// Config 数据库分片配置
type Config struct {
	DbName   string            `yaml:"dbname,omitempty"`   // 数据库名称
	Db       uint32            `yaml:"db,omitempty"`       // 数据库编号
	User     string            `yaml:"user,omitempty"`     // 数据库用户名
	Password string            `yaml:"password,omitempty"` // 数据库密码
	Ip       string            `yaml:"ip,omitempty"`       // 数据库IP地址
	Port     uint32            `yaml:"port,omitempty"`     // 数据库端口
	Prefix   string            `yaml:"prefix,omitempty"`   // key 前缀
	Slaves   map[int32]*Config `yaml:"slaves,omitempty"`   // 从库配置列表
}

type RedisPool struct {
	newFunc  func(*Config) (IClient, error)          // new函数
	pools    map[string]IClient                      // 全局数据库连接池
	virtuals *consistent.StaticHash[string, IClient] // 一致性哈希
}

func NewRedisPool[T IClient](f func(*Config) (T, error)) *RedisPool {
	return &RedisPool{
		newFunc:  func(cfg *Config) (IClient, error) { return f(cfg) },
		pools:    make(map[string]IClient),
		virtuals: consistent.NewStaticHash[string, IClient](150),
	}
}

func (d *RedisPool) Init(globals []*Config, shards []*Config) error {
	// 初始化全局数据库
	for _, dbCfg := range globals {
		cli, err := d.newFunc(dbCfg)
		if err != nil {
			d.Close()
			return err
		}
		d.pools[cli.DbName()] = cli
	}
	// 初始化分片
	for _, dbCfg := range shards {
		cli, err := d.newFunc(dbCfg)
		if err != nil {
			d.Close()
			return err
		}
		d.pools[cli.DbName()] = cli
		if err := d.virtuals.AddNode(cli.DbName(), cli); err != nil {
			return err
		}
	}
	return nil
}

func (d *RedisPool) Close() {
	for _, cli := range d.pools {
		cli.Close()
	}
}

func (d *RedisPool) Get(name string) IClient {
	return d.pools[name]
}

// 通过hash计算分配节点
func (d *RedisPool) GetByHash(seed uint64) IClient {
	return d.virtuals.GetNodeByHash(seed)
}
