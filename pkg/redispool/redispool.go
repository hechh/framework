package redispool

import (
	"time"

	"github.com/hechh/framework/library/consistent"
	"github.com/hechh/framework/library/tplutil"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
)

const (
	HASH   = 1
	STRING = 2
)

type Message interface {
	CloneMessageVT() proto.Message
	MarshalVT() ([]byte, error)
	UnmarshalVT([]byte) error
}

type IClient interface {
	UniqueId() uint32
	Init(cfg *Config) error
	Close() error
	GetRealKey(key string) string
	Run(script *redis.Script, key string, values ...any) (any, error)
	Ping() (string, error)
	Del(keys ...string) (int64, error)
	Exists(key string) (int64, error)
	Expire(key string, expiration time.Duration) (bool, error)
	TTL(key string) (time.Duration, error)
	Get(key string) (string, error)
	Set(key string, val any, expiration time.Duration) error
	SetNX(key string, val any, expiration time.Duration) (bool, error)
	SetEX(key string, val any, expiration time.Duration) error
	Incr(key string) (int64, error)
	IncrBy(key string, val int64) (int64, error)
	Decr(key string) (int64, error)
	DecrBy(key string, value int64) (int64, error)
	MGet(keys ...string) ([]any, error)
	MSet(args ...any) error
	SAdd(key string, members ...any) (int64, error)
	SRem(key string, members ...any) (int64, error)
	SMembers(key string) ([]string, error)
	SIsMember(key string, member any) (bool, error)
	SCard(key string) (int64, error)
	SRandMemberN(key string, count int64) ([]string, error)
	ZAdd(key string, members ...redis.Z) (int64, error)
	ZRem(key string, members ...any) (int64, error)
	ZCard(key string) (int64, error)
	ZScore(key, member string) (float64, error)
	ZRevRange(key string, start, stop int64) ([]string, error)
	ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error)
	ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error)
	ZRevRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error)
	ZRevRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error)
	ZRank(key, member string) (int64, error)
	ZRevRank(key, member string) (int64, error)
	LPush(key string, values ...any) (int64, error)
	RPush(key string, values ...any) (int64, error)
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	LLen(key string) (int64, error)
	LTrim(key string, start, stop int64) error
	LRem(key string, count int64, value any) (int64, error)
	HGet(key string, field string) (string, error)
	HSet(key string, field string, val any) error
	HMGet(key string, fields ...string) ([]any, error)
	HMSet(key string, vals ...any) error
	HDel(key string, fields ...string) (int64, error)
	HExists(key, field string) (bool, error)
	HIncrBy(key, field string, incr int64) (int64, error)
	HLen(key string) (int64, error)
	HSetNX(key, field string, value any) (bool, error)
}

type Value struct {
	Message
	cli   IClient
	class uint32
	key   string
	field string
	times uint32
}

func (d *Value) Client() IClient { return d.cli }
func (d *Value) Type() uint32    { return d.class }
func (d *Value) Key() string     { return d.key }
func (d *Value) Field() string   { return d.field }
func (d *Value) IsChanged() bool { return d.times > 0 }
func (d *Value) Change()         { d.times++ }
func (d *Value) Reset()          { d.times = 0 }
func (d *Value) Get() any        { return d.Message }

func (d *Value) Clone() *Value {
	return &Value{
		Message: d.CloneMessageVT().(Message),
		cli:     d.cli,
		class:   d.class,
		key:     d.key,
		field:   d.field,
	}
}

func NewValue(cli IClient, obj Message, t uint32, args ...string) *Value {
	return &Value{
		Message: obj,
		cli:     cli,
		class:   t,
		key:     tplutil.Index(args, 0, ""),
		field:   tplutil.Index(args, 1, ""),
	}
}

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
	newFunc  func() IClient                          // new函数
	pools    map[string]IClient                      // 全局数据库连接池
	virtuals *consistent.StaticHash[string, IClient] // 一致性哈希
}

func NewRedisPool[T IClient](f func() T) *RedisPool {
	return &RedisPool{
		newFunc:  func() IClient { return f() },
		pools:    make(map[string]IClient),
		virtuals: consistent.NewStaticHash[string, IClient](150),
	}
}

func (d *RedisPool) Init(globals []*Config, shards []*Config) error {
	// 初始化全局数据库
	for _, dbCfg := range globals {
		cli := d.newFunc()
		if err := cli.Init(dbCfg); err != nil {
			d.Close()
			return err
		}
		d.pools[dbCfg.DbName] = cli
	}
	// 初始化分片
	for _, dbCfg := range shards {
		cli := d.newFunc()
		if err := cli.Init(dbCfg); err != nil {
			d.Close()
			return err
		}
		d.pools[dbCfg.DbName] = cli
		if err := d.virtuals.AddNode(dbCfg.DbName, cli); err != nil {
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
