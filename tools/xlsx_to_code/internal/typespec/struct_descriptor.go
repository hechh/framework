package typespec

import (
	"fmt"
	"framework/library/convertor"
	"framework/tools/xlsx_to_code/domain"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type Field struct {
	protoreflect.FieldDescriptor
	Name string
	Type string
}

type Index struct {
	Kind domain.Token
	Name string
	Type string
	List []*Field
}

type StructDescriptor struct {
	protoreflect.MessageType
	Name string
	Data map[string]*Index
	List []*Index
	data map[string]*Field
	list []*Field
}

func NewStructDescriptor(name string, msgType protoreflect.MessageType) *StructDescriptor {
	return &StructDescriptor{
		MessageType: msgType,
		Name:        name,
		Data:        make(map[string]*Index),
		data:        make(map[string]*Field),
	}
}

func (d *StructDescriptor) AddIndex(rule string) {
	strs := strings.Split(rule, ":")
	item := &Index{Name: strings.ReplaceAll(strs[1], ",", "")}
	for _, name := range strings.Split(strs[1], ",") {
		item.List = append(item.List, d.data[name])
	}
	switch strings.ToLower(strs[0]) {
	case "map":
		item.Kind = domain.MAP
		item.Type = fmt.Sprintf("structure.Map%d", len(item.List))
	case "group":
		item.Kind = domain.GROUP
		item.Type = fmt.Sprintf("structure.Group%d", len(item.List))
	}
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructDescriptor) Put(pos int32, name, tname string) {
	item := d.parse(tname)
	item.Name = name
	item.FieldDescriptor = d.Descriptor().Fields().ByName(protoreflect.Name(name))
	d.list = append(d.list, item)
	d.data[item.Name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{Type: item.Type}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{Type: item.Type}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{Type: item.Type}
	}
	return &Field{Type: convertor.Target(str)}
}

func (d *Index) GetArg() string {
	strs := []string{}
	for _, val := range d.List {
		if val.Kind() == protoreflect.EnumKind {
			strs = append(strs, val.Name+" pb."+val.Type)
		} else {
			strs = append(strs, val.Name+" "+val.Type)
		}
	}
	return strings.Join(strs, ", ")
}

func (d *Index) GetType() string {
	strs := []string{}
	for _, val := range d.List {
		if val.Kind() == protoreflect.EnumKind {
			strs = append(strs, "pb."+val.Type)
		} else {
			strs = append(strs, val.Type)
		}
	}
	return strings.Join(strs, ", ")
}

func (d *Index) GetValue(ref string) string {
	strs := []string{}
	for _, val := range d.List {
		if len(ref) > 0 {
			strs = append(strs, ref+"."+val.Name)
		} else {
			strs = append(strs, val.Name)
		}
	}
	return strings.Join(strs, ", ")
}
