package goredis

import (
	"context"
	"fmt"
	"time"

	"github.com/hechh/framework/library/utils"
	"github.com/hechh/framework/pkg/redispool"
	"github.com/redis/go-redis/v9"
)

// Client Redis客户端封装，组合go-redis.Client并添加key前缀支持
type Client struct {
	uuid uint32
	*redis.Client
	cfg *redispool.Config
}

func New() *Client {
	return new(Client)
}

// handleRedisError 将 redis.Nil 转换为 nil，避免上层误判
func handleRedisError(err error) error {
	if err == nil || err == redis.Nil {
		return nil
	}
	return fmt.Errorf("redis error: %w", err)
}

func (d *Client) DbName() string   { return d.cfg.DbName }
func (d *Client) IsSharded() bool  { return d.cfg.IsSharded }
func (d *Client) UniqueId() uint32 { return d.uuid }

func (d *Client) Init(cfg *redispool.Config) error {
	cli := redis.NewClient(&redis.Options{
		ConnMaxIdleTime: 1 * time.Minute, // v9：对应 v8 的 IdleTimeout
		MinIdleConns:    100,
		DB:              int(cfg.Db),
		ReadTimeout:     -1,
		WriteTimeout:    -1,
		Addr:            fmt.Sprintf("%s:%d", cfg.Ip, cfg.Port),
		Username:        cfg.User,
		Password:        cfg.Password,
		OnConnect:       func(ctx context.Context, cn *redis.Conn) error { return nil },
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
		PoolSize:        200,
		ConnMaxLifetime: 0, // v9：对应 v8 的 MaxConnAge（0 表示不过期）
		PoolTimeout:     4 * time.Second,
	})

	// 检查连接是否成功
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := cli.Ping(ctx).Result(); err != nil {
		cli.Close()
		return fmt.Errorf("redis ping: %w", err)
	}
	d.uuid = utils.GetCrc32(fmt.Sprintf("%s-%d", cfg.DbName, cfg.Db))
	d.Client = cli
	d.cfg = cfg
	return nil
}

// Close 关闭连接，将redis.Nil转换为nil
func (d *Client) Close() error {
	err := d.Client.Close()
	return handleRedisError(err)
}

// GetRealKey 获取带前缀的key（供各领域操作使用）
func (c *Client) GetRealKey(key string) string {
	if c.cfg == nil || c.cfg.Prefix == "" {
		return key
	}
	return c.cfg.Prefix + "_" + key
}

// Ctx 返回默认context（供内部使用）
func (*Client) Ctx() context.Context { return context.Background() }

// Pipeline 返回Pipeline
func (d *Client) Pipeline() redis.Pipeliner {
	return d.Client.Pipeline()
}

// TxPipeline 返回事务Pipeline
func (d *Client) TxPipeline() redis.Pipeliner {
	return d.Client.TxPipeline()
}

// Run 执行Lua脚本
func (d *Client) Run(script *redis.Script, key string, values ...any) (any, error) {
	return script.Run(context.Background(), d.Client, []string{d.GetRealKey(key)}, values...).Result()
}

// ClusterNodes 获取集群节点信息
func (d *Client) ClusterNodes() (string, error) {
	flag, err := d.Client.ClusterNodes(context.Background()).Result()
	return flag, handleRedisError(err)
}

// Ping 检查连接
func (d *Client) Ping() (string, error) {
	flag, err := d.Client.Ping(context.Background()).Result()
	return flag, handleRedisError(err)
}

// Publish 发布消息到channel
func (d *Client) Publish(channel string, message any) (int64, error) {
	flag, err := d.Client.Publish(context.Background(), d.GetRealKey(channel), message).Result()
	return flag, handleRedisError(err)
}

// Del 删除keys
func (d *Client) Del(keys ...string) (int64, error) {
	realKeys := make([]string, len(keys))
	for i, key := range keys {
		realKeys[i] = d.GetRealKey(key)
	}
	flag, err := d.Client.Del(d.Ctx(), realKeys...).Result()
	return flag, handleRedisError(err)
}

// Exists 检查key是否存在
func (d *Client) Exists(key string) (int64, error) {
	flag, err := d.Client.Exists(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// Expire 设置key过期时间
func (d *Client) Expire(key string, expiration time.Duration) (bool, error) {
	flag, err := d.Client.Expire(d.Ctx(), d.GetRealKey(key), expiration).Result()
	return flag, handleRedisError(err)
}

// TTL 获取key剩余过期时间
func (d *Client) TTL(key string) (time.Duration, error) {
	ttl, err := d.Client.TTL(d.Ctx(), d.GetRealKey(key)).Result()
	return ttl, handleRedisError(err)
}

// Get 获取string值
func (d *Client) Get(key string) (str string, err error) {
	str, err = d.Client.Get(d.Ctx(), d.GetRealKey(key)).Result()
	return str, handleRedisError(err)
}

// Set 设置string值
func (d *Client) Set(key string, val any, expiration time.Duration) (err error) {
	_, err = d.Client.Set(d.Ctx(), d.GetRealKey(key), val, expiration).Result()
	return handleRedisError(err)
}

// SetNX 不存在key时设置值
func (d *Client) SetNX(key string, val any, expiration time.Duration) (exist bool, err error) {
	exist, err = d.Client.SetNX(d.Ctx(), d.GetRealKey(key), val, expiration).Result()
	return exist, handleRedisError(err)
}

// SetEX 设置值并设置过期时间
func (d *Client) SetEX(key string, val any, expiration time.Duration) (err error) {
	// v9：方法名为 SetEx
	_, err = d.Client.SetEx(d.Ctx(), d.GetRealKey(key), val, expiration).Result()
	return handleRedisError(err)
}

// Incr 自增
func (d *Client) Incr(key string) (ret int64, err error) {
	ret, err = d.Client.Incr(d.Ctx(), d.GetRealKey(key)).Result()
	return ret, handleRedisError(err)
}

// IncrBy 自增指定值
func (d *Client) IncrBy(key string, val int64) (ret int64, err error) {
	ret, err = d.Client.IncrBy(d.Ctx(), d.GetRealKey(key), val).Result()
	return ret, handleRedisError(err)
}

// Decr 自减
func (d *Client) Decr(key string) (int64, error) {
	flag, err := d.Client.Decr(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// DecrBy 自减指定值
func (d *Client) DecrBy(key string, value int64) (int64, error) {
	flag, err := d.Client.DecrBy(d.Ctx(), d.GetRealKey(key), value).Result()
	return flag, handleRedisError(err)
}

// MGet 批量获取string值
func (d *Client) MGet(keys ...string) (rets []any, err error) {
	if len(keys) <= 0 {
		return
	}
	realKeys := make([]string, len(keys))
	for i, k := range keys {
		realKeys[i] = d.GetRealKey(k)
	}
	rets, err = d.Client.MGet(d.Ctx(), realKeys...).Result()
	return rets, handleRedisError(err)
}

// MSet 批量设置string值
func (d *Client) MSet(args ...any) (err error) {
	if len(args) <= 0 {
		return
	}
	realArgs := make([]any, len(args))
	for i, arg := range args {
		if i%2 == 0 {
			realArgs[i] = d.GetRealKey(arg.(string))
		} else {
			realArgs[i] = arg
		}
	}
	_, err = d.Client.MSet(d.Ctx(), realArgs...).Result()
	return handleRedisError(err)
}

// SAdd 添加集合元素
func (d *Client) SAdd(key string, members ...any) (int64, error) {
	flag, err := d.Client.SAdd(d.Ctx(), d.GetRealKey(key), members...).Result()
	return flag, handleRedisError(err)
}

// SRem 删除集合元素
func (d *Client) SRem(key string, members ...any) (int64, error) {
	flag, err := d.Client.SRem(d.Ctx(), d.GetRealKey(key), members...).Result()
	return flag, handleRedisError(err)
}

// SMembers 获取所有集合元素
func (d *Client) SMembers(key string) ([]string, error) {
	flag, err := d.Client.SMembers(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// SIsMember 检查元素是否在集合中
func (d *Client) SIsMember(key string, member any) (bool, error) {
	flag, err := d.Client.SIsMember(d.Ctx(), d.GetRealKey(key), member).Result()
	return flag, handleRedisError(err)
}

// SCard 获取集合元素数量
func (d *Client) SCard(key string) (int64, error) {
	flag, err := d.Client.SCard(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// SRandMemberN 随机获取指定数量的集合元素（不重复）
func (d *Client) SRandMemberN(key string, count int64) ([]string, error) {
	flag, err := d.Client.SRandMemberN(d.Ctx(), d.GetRealKey(key), count).Result()
	return flag, handleRedisError(err)
}

// ZAdd 添加有序集合元素
func (d *Client) ZAdd(key string, members ...redis.Z) (int64, error) {
	flag, err := d.Client.ZAdd(d.Ctx(), d.GetRealKey(key), members...).Result()
	return flag, handleRedisError(err)
}

// ZRemRangeByRank 按排名范围删除有序集合元素
func (d *Client) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	flag, err := d.Client.ZRemRangeByRank(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// ZRem 删除有序集合元素
func (d *Client) ZRem(key string, members ...any) (int64, error) {
	flag, err := d.Client.ZRem(d.Ctx(), d.GetRealKey(key), members...).Result()
	return flag, handleRedisError(err)
}

// ZCard 获取有序集合元素数量
func (d *Client) ZCard(key string) (int64, error) {
	flag, err := d.Client.ZCard(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// ZScore 获取元素分数
func (d *Client) ZScore(key, member string) (float64, error) {
	flag, err := d.Client.ZScore(d.Ctx(), d.GetRealKey(key), member).Result()
	return flag, handleRedisError(err)
}

// ZRange 获取有序集合范围
func (d *Client) ZRange(key string, start, stop int64) ([]string, error) {
	flag, err := d.Client.ZRange(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// ZRevRange 获取有序集合反向范围
func (d *Client) ZRevRange(key string, start, stop int64) ([]string, error) {
	flag, err := d.Client.ZRevRange(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// ZRangeWithScores 获取有序集合范围（带分数）
func (d *Client) ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	flag, err := d.Client.ZRangeWithScores(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// ZRevRangeWithScores 获取有序集合反向范围（带分数）
func (d *Client) ZRevRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	flag, err := d.Client.ZRevRangeWithScores(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// ZRevRangeByScore 按分数范围反向获取有序集合元素
func (d *Client) ZRevRangeByScore(key string, opt *redis.ZRangeBy) ([]string, error) {
	flag, err := d.Client.ZRevRangeByScore(d.Ctx(), d.GetRealKey(key), opt).Result()
	return flag, handleRedisError(err)
}

// ZRevRangeByScoreWithScores 按分数范围反向获取有序集合元素（带分数）
func (d *Client) ZRevRangeByScoreWithScores(key string, opt *redis.ZRangeBy) ([]redis.Z, error) {
	flag, err := d.Client.ZRevRangeByScoreWithScores(d.Ctx(), d.GetRealKey(key), opt).Result()
	return flag, handleRedisError(err)
}

// ZRank 获取元素排名
func (d *Client) ZRank(key, member string) (int64, error) {
	flag, err := d.Client.ZRank(d.Ctx(), d.GetRealKey(key), member).Result()
	return flag, handleRedisError(err)
}

// ZRevRank 获取元素反向排名
func (d *Client) ZRevRank(key, member string) (int64, error) {
	flag, err := d.Client.ZRevRank(d.Ctx(), d.GetRealKey(key), member).Result()
	return flag, handleRedisError(err)
}

// LPush 左侧推入
func (d *Client) LPush(key string, values ...any) (int64, error) {
	flag, err := d.Client.LPush(d.Ctx(), d.GetRealKey(key), values...).Result()
	return flag, handleRedisError(err)
}

// RPush 右侧推入
func (d *Client) RPush(key string, values ...any) (int64, error) {
	flag, err := d.Client.RPush(d.Ctx(), d.GetRealKey(key), values...).Result()
	return flag, handleRedisError(err)
}

// LPop 左侧弹出
func (d *Client) LPop(key string) (string, error) {
	flag, err := d.Client.LPop(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// RPop 右侧弹出
func (d *Client) RPop(key string) (string, error) {
	flag, err := d.Client.RPop(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// LLen 获取列表长度
func (d *Client) LLen(key string) (int64, error) {
	flag, err := d.Client.LLen(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// LRange 获取列表范围
func (d *Client) LRange(key string, start, stop int64) ([]string, error) {
	flag, err := d.Client.LRange(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return flag, handleRedisError(err)
}

// LTrim 截断列表
func (d *Client) LTrim(key string, start, stop int64) error {
	_, err := d.Client.LTrim(d.Ctx(), d.GetRealKey(key), start, stop).Result()
	return handleRedisError(err)
}

// LRem 删除列表元素
func (d *Client) LRem(key string, count int64, value any) (int64, error) {
	flag, err := d.Client.LRem(d.Ctx(), d.GetRealKey(key), count, value).Result()
	return flag, handleRedisError(err)
}

// HGet 获取hash字段值
func (d *Client) HGet(key string, field string) (ret string, err error) {
	ret, err = d.Client.HGet(d.Ctx(), d.GetRealKey(key), field).Result()
	return ret, handleRedisError(err)
}

// HSet 设置hash字段值
func (d *Client) HSet(key string, field string, val any) (err error) {
	_, err = d.Client.HSet(d.Ctx(), d.GetRealKey(key), field, val).Result()
	return handleRedisError(err)
}

// HMGet 批量获取hash字段值
func (d *Client) HMGet(key string, fields ...string) ([]any, error) {
	rets, err := d.Client.HMGet(d.Ctx(), d.GetRealKey(key), fields...).Result()
	return rets, handleRedisError(err)
}

// HMSet 批量设置hash字段值
func (d *Client) HMSet(key string, vals ...any) (err error) {
	// v9：已移除 HMSet，用 HSet(key, field, value, ...) 等价实现
	_, err = d.Client.HSet(d.Ctx(), d.GetRealKey(key), vals...).Result()
	return handleRedisError(err)
}

// HGetAll 获取所有hash字段
func (d *Client) HGetAll(key string) (ret map[string]string, err error) {
	ret, err = d.Client.HGetAll(d.Ctx(), d.GetRealKey(key)).Result()
	return ret, handleRedisError(err)
}

// HDel 删除hash字段
func (d *Client) HDel(key string, fields ...string) (int64, error) {
	flag, err := d.Client.HDel(d.Ctx(), d.GetRealKey(key), fields...).Result()
	return flag, handleRedisError(err)
}

// HExists 检查hash字段是否存在
func (d *Client) HExists(key, field string) (bool, error) {
	flag, err := d.Client.HExists(d.Ctx(), d.GetRealKey(key), field).Result()
	return flag, handleRedisError(err)
}

// HIncrBy hash字段自增
func (d *Client) HIncrBy(key, field string, incr int64) (int64, error) {
	flag, err := d.Client.HIncrBy(d.Ctx(), d.GetRealKey(key), field, incr).Result()
	return flag, handleRedisError(err)
}

// HKeys 获取所有hash字段名
func (d *Client) HKeys(key string) ([]string, error) {
	rets, err := d.Client.HKeys(d.Ctx(), d.GetRealKey(key)).Result()
	return rets, handleRedisError(err)
}

// HLen 获取hash字段数量
func (d *Client) HLen(key string) (int64, error) {
	flag, err := d.Client.HLen(d.Ctx(), d.GetRealKey(key)).Result()
	return flag, handleRedisError(err)
}

// HSetNX hash字段不存在时设置
func (d *Client) HSetNX(key, field string, value any) (bool, error) {
	flag, err := d.Client.HSetNX(d.Ctx(), d.GetRealKey(key), field, value).Result()
	return flag, handleRedisError(err)
}
