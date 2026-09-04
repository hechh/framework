package mockredis

import (
	"testing"
	"time"

	"github.com/hechh/framework/pkg/redispool"
	"github.com/redis/go-redis/v9"
)

// newTestClient 创建并初始化一个 mockredis.Client，测试失败时自动 t.Fatal
func newTestClient(t *testing.T) *Client {
	t.Helper()
	c := New()
	if err := c.Init(&redispool.Config{
		Db:     0,
		Prefix: "test",
	}); err != nil {
		t.Fatalf("Init() 失败: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

// ==================== 基础生命周期 ====================

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() 返回 nil")
	}
	if c.Client != nil {
		t.Fatal("New() 后 Client 应该为 nil")
	}
}

func TestInit(t *testing.T) {
	c := New()
	if err := c.Init(&redispool.Config{
		Db:     0,
		Prefix: "test",
	}); err != nil {
		t.Fatalf("Init() 失败: %v", err)
	}
	defer c.Close()

	if c.Client == nil {
		t.Fatal("Init() 后 Client 不应为 nil")
	}
	if c.miniredis == nil {
		t.Fatal("Init() 后 miniredis 不应为 nil")
	}

	// 验证连接可用
	pong, err := c.Ping()
	if err != nil {
		t.Fatalf("Ping() 失败: %v", err)
	}
	if pong != "PONG" {
		t.Fatalf("Ping() 期望 PONG，得到 %s", pong)
	}
}

func TestInitTwice(t *testing.T) {
	c := newTestClient(t)
	// 第二次 Init 应该正常（重新连接）
	if err := c.Init(&redispool.Config{
		Db:     1,
		Prefix: "test2",
	}); err != nil {
		t.Fatalf("第二次 Init() 失败: %v", err)
	}
}

func TestClose(t *testing.T) {
	c := newTestClient(t)
	if err := c.Close(); err != nil {
		t.Fatalf("Close() 失败: %v", err)
	}
	// 重复关闭不应 panic
	if err := c.Close(); err != nil {
		t.Fatalf("重复 Close() 失败: %v", err)
	}
}

// ==================== String 操作 ====================

func TestSetGet(t *testing.T) {
	c := newTestClient(t)

	if err := c.Set("key1", "value1", 0); err != nil {
		t.Fatalf("Set() 失败: %v", err)
	}
	val, err := c.Get("key1")
	if err != nil {
		t.Fatalf("Get() 失败: %v", err)
	}
	if val != "value1" {
		t.Fatalf("Get() 期望 value1，得到 %s", val)
	}
}

func TestGetNil(t *testing.T) {
	c := newTestClient(t)

	val, err := c.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get() 不存在的 key 应返回 nil error，得到: %v", err)
	}
	if val != "" {
		t.Fatalf("Get() 不存在的 key 应返回空字符串，得到: %s", val)
	}
}

func TestSetNX(t *testing.T) {
	c := newTestClient(t)

	ok, err := c.SetNX("nx_key", "val1", 0)
	if err != nil {
		t.Fatalf("SetNX() 失败: %v", err)
	}
	if !ok {
		t.Fatal("SetNX() 新 key 应返回 true")
	}

	ok, err = c.SetNX("nx_key", "val2", 0)
	if err != nil {
		t.Fatalf("第二次 SetNX() 失败: %v", err)
	}
	if ok {
		t.Fatal("SetNX() 已存在的 key 应返回 false")
	}

	val, _ := c.Get("nx_key")
	if val != "val1" {
		t.Fatalf("Get() 期望 val1，得到 %s", val)
	}
}

func TestSetEX(t *testing.T) {
	c := newTestClient(t)

	if err := c.SetEX("ex_key", "ex_val", time.Hour); err != nil {
		t.Fatalf("SetEX() 失败: %v", err)
	}
	val, err := c.Get("ex_key")
	if err != nil {
		t.Fatalf("Get() 失败: %v", err)
	}
	if val != "ex_val" {
		t.Fatalf("Get() 期望 ex_val，得到 %s", val)
	}
	ttl, err := c.TTL("ex_key")
	if err != nil {
		t.Fatalf("TTL() 失败: %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("TTL() 应大于 0，得到 %v", ttl)
	}
}

func TestIncrDecr(t *testing.T) {
	c := newTestClient(t)

	n, err := c.Incr("counter")
	if err != nil {
		t.Fatalf("Incr() 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("Incr() 期望 1，得到 %d", n)
	}

	n, err = c.IncrBy("counter", 5)
	if err != nil {
		t.Fatalf("IncrBy() 失败: %v", err)
	}
	if n != 6 {
		t.Fatalf("IncrBy() 期望 6，得到 %d", n)
	}

	n, err = c.Decr("counter")
	if err != nil {
		t.Fatalf("Decr() 失败: %v", err)
	}
	if n != 5 {
		t.Fatalf("Decr() 期望 5，得到 %d", n)
	}

	n, err = c.DecrBy("counter", 3)
	if err != nil {
		t.Fatalf("DecrBy() 失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("DecrBy() 期望 2，得到 %d", n)
	}
}

func TestMGetMSet(t *testing.T) {
	c := newTestClient(t)

	if err := c.MSet("k1", "v1", "k2", "v2", "k3", "v3"); err != nil {
		t.Fatalf("MSet() 失败: %v", err)
	}

	vals, err := c.MGet("k1", "k2", "k3")
	if err != nil {
		t.Fatalf("MGet() 失败: %v", err)
	}
	if len(vals) != 3 {
		t.Fatalf("MGet() 期望 3 个结果，得到 %d", len(vals))
	}
	for i, expected := range []string{"v1", "v2", "v3"} {
		if vals[i] != expected {
			t.Fatalf("MGet()[%d] 期望 %s，得到 %v", i, expected, vals[i])
		}
	}
}

// ==================== Hash 操作 ====================

func TestHash(t *testing.T) {
	c := newTestClient(t)

	// HSet / HGet
	if err := c.HSet("hash1", "field1", "val1"); err != nil {
		t.Fatalf("HSet() 失败: %v", err)
	}
	val, err := c.HGet("hash1", "field1")
	if err != nil {
		t.Fatalf("HGet() 失败: %v", err)
	}
	if val != "val1" {
		t.Fatalf("HGet() 期望 val1，得到 %s", val)
	}

	// HGet 不存在的字段
	val, err = c.HGet("hash1", "nonexistent")
	if err != nil {
		t.Fatalf("HGet() 不存在的字段应返回 nil error，得到: %v", err)
	}
	if val != "" {
		t.Fatalf("HGet() 不存在的字段应返回空字符串，得到 %s", val)
	}

	// HMSet / HMGet
	if err := c.HMSet("hash1", "f1", "a", "f2", "b", "f3", "c"); err != nil {
		t.Fatalf("HMSet() 失败: %v", err)
	}
	vals, err := c.HMGet("hash1", "f1", "f2", "f3")
	if err != nil {
		t.Fatalf("HMGet() 失败: %v", err)
	}
	if len(vals) != 3 {
		t.Fatalf("HMGet() 期望 3 个结果，得到 %d", len(vals))
	}

	// HGetAll
	all, err := c.HGetAll("hash1")
	if err != nil {
		t.Fatalf("HGetAll() 失败: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("HGetAll() 不应为空")
	}
	if all["f1"] != "a" || all["f2"] != "b" || all["f3"] != "c" {
		t.Fatalf("HGetAll() 结果不正确: %v", all)
	}

	// HExists
	exists, err := c.HExists("hash1", "f1")
	if err != nil {
		t.Fatalf("HExists() 失败: %v", err)
	}
	if !exists {
		t.Fatal("HExists() 存在的字段应返回 true")
	}
	exists, err = c.HExists("hash1", "missing")
	if err != nil {
		t.Fatalf("HExists() 失败: %v", err)
	}
	if exists {
		t.Fatal("HExists() 不存在的字段应返回 false")
	}

	// HLen
	hlen, err := c.HLen("hash1")
	if err != nil {
		t.Fatalf("HLen() 失败: %v", err)
	}
	if hlen != 4 {
		t.Fatalf("HLen() 期望 4，得到 %d", hlen)
	}

	// HKeys
	keys, err := c.HKeys("hash1")
	if err != nil {
		t.Fatalf("HKeys() 失败: %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("HKeys() 不应为空")
	}

	// HDel
	deleted, err := c.HDel("hash1", "f1", "f2")
	if err != nil {
		t.Fatalf("HDel() 失败: %v", err)
	}
	if deleted != 2 {
		t.Fatalf("HDel() 期望删除 2 个，得到 %d", deleted)
	}

	// HIncrBy
	n, err := c.HIncrBy("hash1", "counter", 10)
	if err != nil {
		t.Fatalf("HIncrBy() 失败: %v", err)
	}
	if n != 10 {
		t.Fatalf("HIncrBy() 期望 10，得到 %d", n)
	}

	// HSetNX
	ok, err := c.HSetNX("hash1", "new_field", "new_val")
	if err != nil {
		t.Fatalf("HSetNX() 失败: %v", err)
	}
	if !ok {
		t.Fatal("HSetNX() 新字段应返回 true")
	}
	ok, err = c.HSetNX("hash1", "new_field", "other")
	if err != nil {
		t.Fatalf("HSetNX() 重复字段失败: %v", err)
	}
	if ok {
		t.Fatal("HSetNX() 已存在的字段应返回 false")
	}
}

// ==================== List 操作 ====================

func TestList(t *testing.T) {
	c := newTestClient(t)

	// LPush / LLen
	n, err := c.LPush("list1", "c", "b", "a")
	if err != nil {
		t.Fatalf("LPush() 失败: %v", err)
	}
	if n != 3 {
		t.Fatalf("LPush() 期望 3，得到 %d", n)
	}
	llen, err := c.LLen("list1")
	if err != nil {
		t.Fatalf("LLen() 失败: %v", err)
	}
	if llen != 3 {
		t.Fatalf("LLen() 期望 3，得到 %d", llen)
	}

	// RPush
	n, err = c.RPush("list1", "d", "e")
	if err != nil {
		t.Fatalf("RPush() 失败: %v", err)
	}
	if n != 5 {
		t.Fatalf("RPush() 期望 5，得到 %d", n)
	}

	// LRange
	vals, err := c.LRange("list1", 0, -1)
	if err != nil {
		t.Fatalf("LRange() 失败: %v", err)
	}
	expected := []string{"a", "b", "c", "d", "e"}
	if len(vals) != len(expected) {
		t.Fatalf("LRange() 期望 %d 个元素，得到 %d", len(expected), len(vals))
	}
	for i := range expected {
		if vals[i] != expected[i] {
			t.Fatalf("LRange()[%d] 期望 %s，得到 %s", i, expected[i], vals[i])
		}
	}

	// LPop
	val, err := c.LPop("list1")
	if err != nil {
		t.Fatalf("LPop() 失败: %v", err)
	}
	if val != "a" {
		t.Fatalf("LPop() 期望 a，得到 %s", val)
	}

	// RPop
	val, err = c.RPop("list1")
	if err != nil {
		t.Fatalf("RPop() 失败: %v", err)
	}
	if val != "e" {
		t.Fatalf("RPop() 期望 e，得到 %s", val)
	}

	// LTrim
	if err := c.LTrim("list1", 0, 0); err != nil {
		t.Fatalf("LTrim() 失败: %v", err)
	}
	llen, _ = c.LLen("list1")
	if llen != 1 {
		t.Fatalf("LTrim() 后 LLen 期望 1，得到 %d", llen)
	}

	// LRem
	n, err = c.LRem("list1", 1, "b")
	if err != nil {
		t.Fatalf("LRem() 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("LRem() 期望删除 1 个，得到 %d", n)
	}
}

// ==================== Set 操作 ====================

func TestSet(t *testing.T) {
	c := newTestClient(t)

	// SAdd
	n, err := c.SAdd("set1", "a", "b", "c", "a")
	if err != nil {
		t.Fatalf("SAdd() 失败: %v", err)
	}
	// miniredis 的 SAdd 返回实际添加数量（不含重复）
	if n != 3 {
		t.Fatalf("SAdd() 期望 3，得到 %d", n)
	}

	// SCard
	card, err := c.SCard("set1")
	if err != nil {
		t.Fatalf("SCard() 失败: %v", err)
	}
	if card != 3 {
		t.Fatalf("SCard() 期望 3，得到 %d", card)
	}

	// SIsMember
	ok, err := c.SIsMember("set1", "a")
	if err != nil {
		t.Fatalf("SIsMember() 失败: %v", err)
	}
	if !ok {
		t.Fatal("SIsMember() 存在的成员应返回 true")
	}
	ok, err = c.SIsMember("set1", "z")
	if err != nil {
		t.Fatalf("SIsMember() 失败: %v", err)
	}
	if ok {
		t.Fatal("SIsMember() 不存在的成员应返回 false")
	}

	// SMembers
	members, err := c.SMembers("set1")
	if err != nil {
		t.Fatalf("SMembers() 失败: %v", err)
	}
	if len(members) != 3 {
		t.Fatalf("SMembers() 期望 3 个元素，得到 %d", len(members))
	}

	// SRandMemberN
	randoms, err := c.SRandMemberN("set1", 2)
	if err != nil {
		t.Fatalf("SRandMemberN() 失败: %v", err)
	}
	if len(randoms) != 2 {
		t.Fatalf("SRandMemberN() 期望 2 个元素，得到 %d", len(randoms))
	}

	// SRem
	n, err = c.SRem("set1", "a", "b")
	if err != nil {
		t.Fatalf("SRem() 失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("SRem() 期望删除 2 个，得到 %d", n)
	}
}

// ==================== ZSet 操作 ====================

func TestZSet(t *testing.T) {
	c := newTestClient(t)

	// ZAdd
	n, err := c.ZAdd("zset1",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	if err != nil {
		t.Fatalf("ZAdd() 失败: %v", err)
	}
	if n != 3 {
		t.Fatalf("ZAdd() 期望 3，得到 %d", n)
	}

	// ZCard
	card, err := c.ZCard("zset1")
	if err != nil {
		t.Fatalf("ZCard() 失败: %v", err)
	}
	if card != 3 {
		t.Fatalf("ZCard() 期望 3，得到 %d", card)
	}

	// ZScore
	score, err := c.ZScore("zset1", "b")
	if err != nil {
		t.Fatalf("ZScore() 失败: %v", err)
	}
	if score != 2 {
		t.Fatalf("ZScore() 期望 2，得到 %f", score)
	}

	// ZRange
	vals, err := c.ZRange("zset1", 0, -1)
	if err != nil {
		t.Fatalf("ZRange() 失败: %v", err)
	}
	expected := []string{"a", "b", "c"}
	if len(vals) != len(expected) {
		t.Fatalf("ZRange() 期望 %d 个元素，得到 %d", len(expected), len(vals))
	}
	for i := range expected {
		if vals[i] != expected[i] {
			t.Fatalf("ZRange()[%d] 期望 %s，得到 %s", i, expected[i], vals[i])
		}
	}

	// ZRevRange
	revVals, err := c.ZRevRange("zset1", 0, -1)
	if err != nil {
		t.Fatalf("ZRevRange() 失败: %v", err)
	}
	if len(revVals) != 3 {
		t.Fatalf("ZRevRange() 期望 3 个元素，得到 %d", len(revVals))
	}
	if revVals[0] != "c" {
		t.Fatalf("ZRevRange() 第一个元素期望 c，得到 %s", revVals[0])
	}

	// ZRangeWithScores
	withScores, err := c.ZRangeWithScores("zset1", 0, -1)
	if err != nil {
		t.Fatalf("ZRangeWithScores() 失败: %v", err)
	}
	if len(withScores) != 3 {
		t.Fatalf("ZRangeWithScores() 期望 3 个元素，得到 %d", len(withScores))
	}

	// ZRevRangeWithScores
	revWithScores, err := c.ZRevRangeWithScores("zset1", 0, -1)
	if err != nil {
		t.Fatalf("ZRevRangeWithScores() 失败: %v", err)
	}
	if len(revWithScores) != 3 {
		t.Fatalf("ZRevRangeWithScores() 期望 3 个元素，得到 %d", len(revWithScores))
	}

	// ZRank / ZRevRank
	rank, err := c.ZRank("zset1", "a")
	if err != nil {
		t.Fatalf("ZRank() 失败: %v", err)
	}
	if rank != 0 {
		t.Fatalf("ZRank('a') 期望 0，得到 %d", rank)
	}
	revRank, err := c.ZRevRank("zset1", "a")
	if err != nil {
		t.Fatalf("ZRevRank() 失败: %v", err)
	}
	if revRank != 2 {
		t.Fatalf("ZRevRank('a') 期望 2，得到 %d", revRank)
	}

	// ZRem
	n, err = c.ZRem("zset1", "a")
	if err != nil {
		t.Fatalf("ZRem() 失败: %v", err)
	}
	if n != 1 {
		t.Fatalf("ZRem() 期望删除 1 个，得到 %d", n)
	}

	// ZRemRangeByRank
	// 先添加一些元素
	_, _ = c.ZAdd("zset1", redis.Z{Score: 10, Member: "x"}, redis.Z{Score: 20, Member: "y"}, redis.Z{Score: 30, Member: "z"})
	n, err = c.ZRemRangeByRank("zset1", 0, 1)
	if err != nil {
		t.Fatalf("ZRemRangeByRank() 失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("ZRemRangeByRank() 期望删除 2 个，得到 %d", n)
	}

	// ZRevRangeByScore
	_, _ = c.ZAdd("zset2",
		redis.Z{Score: 1, Member: "a"},
		redis.Z{Score: 2, Member: "b"},
		redis.Z{Score: 3, Member: "c"},
	)
	rangeByScore, err := c.ZRevRangeByScore("zset2", &redis.ZRangeBy{Min: "1", Max: "3"})
	if err != nil {
		t.Fatalf("ZRevRangeByScore() 失败: %v", err)
	}
	if len(rangeByScore) != 3 {
		t.Fatalf("ZRevRangeByScore() 期望 3 个元素，得到 %d", len(rangeByScore))
	}

	// ZRevRangeByScoreWithScores
	rangeByScoreWithScores, err := c.ZRevRangeByScoreWithScores("zset2", &redis.ZRangeBy{Min: "1", Max: "3"})
	if err != nil {
		t.Fatalf("ZRevRangeByScoreWithScores() 失败: %v", err)
	}
	if len(rangeByScoreWithScores) != 3 {
		t.Fatalf("ZRevRangeByScoreWithScores() 期望 3 个元素，得到 %d", len(rangeByScoreWithScores))
	}
}

// ==================== Key 操作 ====================

func TestDelExists(t *testing.T) {
	c := newTestClient(t)

	_ = c.Set("k1", "v1", 0)
	_ = c.Set("k2", "v2", 0)

	exists, err := c.Exists("k1")
	if err != nil {
		t.Fatalf("Exists() 失败: %v", err)
	}
	if exists != 1 {
		t.Fatalf("Exists() 期望 1，得到 %d", exists)
	}

	n, err := c.Del("k1", "k2")
	if err != nil {
		t.Fatalf("Del() 失败: %v", err)
	}
	if n != 2 {
		t.Fatalf("Del() 期望删除 2 个，得到 %d", n)
	}

	exists, _ = c.Exists("k1")
	if exists != 0 {
		t.Fatalf("Exists() 删除后期望 0，得到 %d", exists)
	}
}

func TestExpireTTL(t *testing.T) {
	c := newTestClient(t)

	_ = c.Set("tmp", "val", 0)

	ok, err := c.Expire("tmp", time.Second)
	if err != nil {
		t.Fatalf("Expire() 失败: %v", err)
	}
	if !ok {
		t.Fatal("Expire() 应返回 true")
	}

	ttl, err := c.TTL("tmp")
	if err != nil {
		t.Fatalf("TTL() 失败: %v", err)
	}
	if ttl <= 0 || ttl > time.Second {
		t.Fatalf("TTL() 期望 0~1s，得到 %v", ttl)
	}

	// 不存在的 key
	ok, err = c.Expire("nonexistent", time.Second)
	if err != nil {
		t.Fatalf("Expire() 不存在的 key 失败: %v", err)
	}
	if ok {
		t.Fatal("Expire() 不存在的 key 应返回 false")
	}

	ttl, err = c.TTL("nonexistent")
	if err != nil {
		t.Fatalf("TTL() 不存在的 key 失败: %v", err)
	}
	if ttl != -2 {
		t.Fatalf("TTL() 不存在的 key 期望 -2，得到 %v", ttl)
	}
}

// ==================== 其他操作 ====================

func TestCtx(t *testing.T) {
	c := newTestClient(t)
	ctx := c.Ctx()
	if ctx == nil {
		t.Fatal("Ctx() 返回 nil")
	}
}

func TestPipeline(t *testing.T) {
	c := newTestClient(t)

	pipe := c.Pipeline()
	if pipe == nil {
		t.Fatal("Pipeline() 返回 nil")
	}
}

func TestTxPipeline(t *testing.T) {
	c := newTestClient(t)

	pipe := c.TxPipeline()
	if pipe == nil {
		t.Fatal("TxPipeline() 返回 nil")
	}
}

func TestGetRealKey(t *testing.T) {
	c := newTestClient(t)

	// mockredis 固定使用 "mock" 前缀
	realKey := c.GetRealKey("mykey")
	expected := "mock_mykey"
	if realKey != expected {
		t.Fatalf("GetRealKey() 期望 %s，得到 %s", expected, realKey)
	}
}

func TestClusterNodes(t *testing.T) {
	c := newTestClient(t)
	// miniredis 非集群模式，期望返回错误或空
	_, err := c.ClusterNodes()
	if err != nil {
		// 允许错误（miniredis 不支持集群）
		t.Logf("ClusterNodes() 返回预期错误: %v", err)
	}
}

func TestPublish(t *testing.T) {
	c := newTestClient(t)

	n, err := c.Publish("channel1", "message1")
	if err != nil {
		t.Fatalf("Publish() 失败: %v", err)
	}
	if n != 0 {
		t.Logf("Publish() 返回订阅者数量: %d", n)
	}
}

func TestRun(t *testing.T) {
	c := newTestClient(t)

	_ = c.Set("foo", "bar", 0)
	script := redis.NewScript("return redis.call('GET', KEYS[1])")
	result, err := c.Run(script, "foo")
	if err != nil {
		t.Fatalf("Run() 失败: %v", err)
	}
	if result != "bar" {
		t.Fatalf("Run() 期望 bar，得到 %v", result)
	}
}

// ==================== 并发安全 ====================

func TestConcurrentAccess(t *testing.T) {
	c := newTestClient(t)

	done := make(chan struct{})
	const goroutines = 20

	for i := 0; i < goroutines; i++ {
		go func(n int) {
			key := "concurrent_key"
			_ = c.Set(key, n, 0)
			val, _ := c.Get(key)
			_, _ = c.Del(key)
			_ = val
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}
}
