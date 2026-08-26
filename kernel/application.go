package kernel

import (
	"os"
	"os/signal"
	"syscall"
)

type IComponent interface {
	Init(map[string]any) error
	Close()
}

type Application struct {
	list []IComponent
}

func NewApplication() *Application {
	return &Application{}
}

func (d *Application) Register(c IComponent) {
	d.list = append(d.list, c)
}

func (d *Application) Init(data map[string]any) error {
	for _, comp := range d.list {
		if err := comp.Init(data); err != nil {
			return err
		}
	}
	return nil
}

func (d *Application) Close() {
	for j := len(d.list) - 1; j >= 0; j-- {
		d.list[j].Close()
	}
}

func (d *Application) Run(sigs ...os.Signal) {
	sigs = append(sigs, syscall.SIGABRT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)
	<-sigChan
	d.Close()
}
