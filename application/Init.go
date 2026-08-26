package application

import (
	"os"
)

var (
	app *Application = NewApplication()
)

func Register(c IComponent) {
	app.Register(c)
}

func Init(data map[string]any) error {
	return app.Init(data)
}

func Close() {
	app.Close()
}

func Run(sigs ...os.Signal) {
	app.Run(sigs...)
}
