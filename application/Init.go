package application

import (
	"os"

	"github.com/hechh/framework/library/fileutil"
)

var (
	app *Application = NewApplication()
)

func Register(c IComponent) {
	app.Register(c)
}

func Init(filename string) error {
	data := make(map[string]any)
	if err := fileutil.LoadYaml(filename, data); err != nil {
		return err
	}
	return app.Init(data)
}

func Close() {
	app.Close()
}

func Run(sigs ...os.Signal) {
	app.Run(sigs...)
}
