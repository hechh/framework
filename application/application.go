package application

import (
	"os"
	"os/signal"
	"syscall"
)

type IComponent interface {
	Init() error
	Close()
}

func NewApplication() *Application {
	return &Application{}
}

type Application struct {
	list []IComponent
}

func (d *Application) Register(c IComponent) {
	d.list = append(d.list, c)
}

func (d *Application) Init() error {
	for _, comp := range d.list {
		if err := comp.Init(); err != nil {
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
