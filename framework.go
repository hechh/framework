package framework

import (
	"os"

	"github.com/hechh/framework/service"
)

var (
	object = service.New()
)

func Register(c any) {
	object.Register(c)
}

func Init() error {
	return object.Init()
}

func Close() {
	object.Close()
}

func Run(sigs ...os.Signal) {
	object.Run(sigs...)
}

func GetApp() *service.Service {
	return object
}
