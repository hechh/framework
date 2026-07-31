package context

import "github.com/hechh/framework/define"

// 编译期断言：Cache 实现 define.ICache 接口
var _ define.ICache = (*Cache)(nil)

type Value struct {
	value any
	mask  uint32
	times uint32
}

type Cache struct {
	values map[string]*Value
}

// NewCache 创建空缓存
func NewCache() *Cache {
	return &Cache{values: make(map[string]*Value)}
}

// SetCache 写入缓存值，flag 记录数据来源标志（已存在则更新值并覆盖标志）
func (d *Cache) SetCache(key string, value any, flag uint32) {
	if val, ok := d.values[key]; ok {
		val.value = value
		val.mask = flag
		return
	}
	d.values[key] = &Value{value: value, mask: flag}
}

// GetCache 读取缓存值
func (d *Cache) GetCache(key string) (any, bool) {
	if v, ok := d.values[key]; ok {
		return v.value, ok
	}
	return nil, false
}

// IsChanged 判断 key 是否被标记为已变更
func (d *Cache) IsChanged(key string) bool {
	if v, ok := d.values[key]; ok {
		return v.times > 0
	}
	return false
}

// Change 标记 key 为已变更（计数累加，供持久化层判定脏数据）
func (d *Cache) Change(key string) {
	if v, ok := d.values[key]; ok {
		v.times++
	}
}

// Reset 清除 key 的变更标记
func (d *Cache) Reset(key string) {
	if v, ok := d.values[key]; ok {
		v.times = 0
	}
}
