package domain

import "framework/define"

type FilterFunc func(define.IRouter) bool                 // 过滤函数
type NewFunc func(idType int32, id uint64) define.IRouter // 创建路由函数
