package cache

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/hechh/framework/library/datetime"
	"github.com/hechh/framework/pkg/mlog"
)

type MigrateData struct {
	Uid        uint64        `json:"uid,omitempty"`
	OldDbName  string        `json:"OldDbName,omitempty"`
	NewDbName  string        `json:"NewDbName,omitempty"`
	parent     *MigrateCache `json:"-"`
	updateTime atomic.Int64  `json:"-"`
	ttlMs      int64         `json:"-"`
}

func (d *MigrateData) IsEnable() bool {
	return true
}

func (d *MigrateData) GetTTL() int64 {
	return d.ttlMs
}

func (d *MigrateData) GetExpire() int64 {
	return d.updateTime.Load() + d.ttlMs
}

func (d *MigrateData) Refresh(val int64) {
	d.updateTime.Store(val)
}

func (d *MigrateData) Call() {
	if d.parent != nil {
		d.parent.Remove(d.Uid)
	}
}

type shardData struct {
	data  map[uint64]*MigrateData // 迁移数据
	mutex sync.RWMutex            // 读写锁
}

type MigrateCache struct {
	shards []*shardData // 分片数据
}

func NewMigrateCache() *MigrateCache {
	shards := make([]*shardData, 0, 256)
	for range 256 {
		shards = append(shards, &shardData{
			data: make(map[uint64]*MigrateData),
		})
	}
	return &MigrateCache{
		shards: shards,
	}
}

func (d *MigrateCache) Get(uid uint64) *MigrateData {
	shard := d.shards[uid%256]
	shard.mutex.RLock()
	item, ok := shard.data[uid]
	shard.mutex.RUnlock()
	if ok {
		item.updateTime.Store(datetime.NowUnixMilli())
	}
	return item
}

func (d *MigrateCache) Add(val *MigrateData) {
	val.parent = d
	val.updateTime.Store(datetime.NowUnixMilli())
	val.ttlMs = int64(15 * time.Minute / time.Millisecond)

	shard := d.shards[val.Uid%256]
	shard.mutex.Lock()
	shard.data[val.Uid] = val
	shard.mutex.Unlock()
}

func (d *MigrateCache) Remove(uid uint64) {
	shard := d.shards[uid%256]
	shard.mutex.Lock()
	item, ok := shard.data[uid]
	if ok {
		delete(shard.data, uid)
	}
	shard.mutex.Unlock()
	if ok {
		mlog.Tracef("migrate data uid:%d, old:%s, new:%s", item.Uid, item.OldDbName, item.NewDbName)
	}
}
