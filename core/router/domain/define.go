package domain

import "framework/core/define"

type NewFunc func(uint32, uint64, int64) define.IRouter // 创建路由函数
