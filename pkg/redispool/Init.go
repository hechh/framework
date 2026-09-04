package redispool

var object *RedisPool

func SetObject(oj *RedisPool) {
	object = oj
}

// GetByName 根据名称获取Redis客户端
func Get(name string) IClient {
	if object != nil {
		return object.Get(name)
	}
	return nil
}

func GetByHash(seed uint64) IClient {
	if object != nil {
		return object.GetByHash(seed)
	}
	return nil
}
