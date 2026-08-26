package dbpool

import (
	"github.com/hechh/framework/pkg/dbpool/internal/registry"
)

var (
	object *DbPool
)

func SetObject(oj *DbPool) {
	object = oj
}

func Register(name string, datas ...any) {
	registry.Register(name, datas...)
}

func Get(name string) IClient {
	if object != nil {
		return object.Get(name)
	}
	return nil
}
