package kafka

var (
	object *Kafka
)

func SetObject(oj *Kafka) {
	object = oj
}

// GetObject 获取全局 Kafka 实例（用于业务模块订阅主题/发送消息）。
// 在 kafka 组件初始化完成后可用，未初始化时返回 nil。
func GetObject() *Kafka {
	return object
}
