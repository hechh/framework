package redispool

import (
	"time"

	"github.com/hechh/framework/define"
	"github.com/hechh/framework/library/safe"
	"github.com/hechh/framework/library/tplutil"
	"github.com/redis/go-redis/v9"
)

const (
	HASH   = 1
	STRING = 2
)

type IClient interface {
	Close() error
	DbName() string
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
	data     define.Message
	cli      IClient
	dataType uint32
	key      string
	field    string
	times    uint32
}

func NewValue(cli IClient, obj define.Message, t uint32, args ...string) *Value {
	return &Value{
		data:     obj,
		cli:      cli,
		dataType: t,
		key:      tplutil.Index(args, 0, ""),
		field:    tplutil.Index(args, 1, ""),
	}
}

func (d *Value) SetClient(cli IClient) { d.cli = cli }
func (d *Value) GetClient() IClient    { return d.cli }
func (d *Value) GetDataType() uint32   { return d.dataType }
func (d *Value) GetKey() string        { return d.key }
func (d *Value) GetField() string      { return d.field }
func (d *Value) GetGroupId() string {
	if d.dataType == HASH {
		return d.cli.DbName() + d.key
	}
	return d.cli.DbName()
}
func (d *Value) Unmarshal(val any) error {
	switch vv := val.(type) {
	case string:
		return d.data.UnmarshalVT(safe.StringToBytes(vv))
	case []byte:
		return d.data.UnmarshalVT(vv)
	default:
		return d.data.UnmarshalVT(nil)
	}
}
func (d *Value) Marshal() ([]byte, error) {
	return d.data.MarshalVT()
}

// 实现define.IValue
func (d *Value) Get() any        { return d.data }
func (d *Value) IsChanged() bool { return d.times > 0 }
func (d *Value) Change()         { d.times++ }
func (d *Value) Reset()          { d.times = 0 }
func (d *Value) Clone() define.IValue {
	return &Value{
		data:     d.data.CloneMessageVT().(define.Message),
		cli:      d.cli,
		dataType: d.dataType,
		key:      d.key,
		field:    d.field,
	}
}
