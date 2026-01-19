package entity

import (
	"reflect"

	"github.com/hechh/framework"
)

type Base struct {
	framework.ISerialize
	name  string
	crc32 uint32
}

func NewBase(en framework.ISerialize, name string, f reflect.Value) *Base {
	if len(name) <= 0 {
		name = framework.ParseActorFunc(f)
	}
	return &Base{
		ISerialize: en,
		name:       name,
		crc32:      framework.GetCrc32(name),
	}
}

func (d *Base) GetCrc32() uint32 {
	return d.crc32
}

func (d *Base) GetName() string {
	return d.name
}
