package application

import "os"

var (
	app *Application = &Application{}
)

func Register(c IComponent) {
	app.Register(c)
}

func Init() error {
	return app.Init()
}

func Close() {
	app.Close()
}

func Run(sigs ...os.Signal) {
	app.Run(sigs...)
}
