package redispool

import (
	"fmt"

	"github.com/hechh/framework/library/consistent"
	"github.com/hechh/framework/pkg/mlog"
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
	if len(globals) == 0 && len(shards) == 0 {
		return fmt.Errorf("redis配置为空：globals 与 shards 均为空，无法初始化")
	}
	if len(shards) == 0 {
		mlog.Warnf("[redispool] shards 配置为空：GetByHash 将返回 nil，按 uid 路由的数据无法读写")
	}

	// 初始化全局数据库
	for _, dbCfg := range globals {
		if err := d.add(dbCfg, false); err != nil {
			d.Close()
			return err
		}
	}
	// 初始化分片
	for _, dbCfg := range shards {
		if err := d.add(dbCfg, true); err != nil {
			d.Close()
			return err
		}
	}

	// 构建结束，进入只读阶段：StaticHash 的无锁读以 Freeze 为安全前提，
	// 不 Freeze 则运行期误调 AddNode 会与读路径并发改 hashRing（切片头撕裂/路由错乱）
	d.virtuals.Freeze()
	return nil
}

// add 创建并注册一个连接，shard 为 true 时同时加入一致性哈希环。
// dbname 同时充当连接池键、哈希环节点 id、批量写分组键与迁移的"同库"判断依据，
// 重名会让连接被静默覆盖（旧连接泄漏）并按错库路由数据，因此直接拒绝。
func (d *RedisPool) add(dbCfg *Config, shard bool) error {
	if _, ok := d.pools[dbCfg.DbName]; ok {
		return fmt.Errorf("redis dbname(%s) 重复配置：dbname 是连接唯一标识（连接池键/哈希环节点/同库判断），重名会导致连接覆盖与数据落错库", dbCfg.DbName)
	}
	cli, err := d.newFunc(dbCfg)
	if err != nil {
		return err
	}
	// 先注册再建环：AddNode 失败时连接留在池中，由调用方的 Close 统一释放，避免连接泄漏
	d.pools[cli.DbName()] = cli
	if !shard {
		return nil
	}
	if err := d.virtuals.AddNode(cli.DbName(), cli); err != nil {
		return err
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
