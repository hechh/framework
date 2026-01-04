package typespec

import (
	"fmt"
	"framework/library/convertor"
	"sort"
	"strings"
)

type Field struct {
	Type     string
	Name     string
	Position int32
	Desc     string
}
type StructDescriptor struct {
	Name string
	List []*Field
	Data map[string]*Field
}

func NewStructDescriptor(Name string) *StructDescriptor {
	return &StructDescriptor{
		Name: Name,
		Data: make(map[string]*Field),
	}
}

func (d *StructDescriptor) Put(pos int32, Name, tname, Desc string) {
	item := d.parse(tname)
	item.Name = Name
	item.Position = pos
	item.Desc = Desc
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{Type: "repeated " + item.Type}
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

func (d *StructDescriptor) String() string {
	sort.Slice(d.List, func(i int, j int) bool {
		return d.List[i].Position < d.List[j].Position
	})
	strs := []string{}
	for _, item := range d.List {
		strs = append(strs, fmt.Sprintf("\t%s %s = %d;\t// %s",
			item.Type, item.Name, item.Position, item.Desc))
	}
	return fmt.Sprintf("message %s {\n%s\n}\n\nmessage %sAry {\nrepeated %s Ary = 1;\n}\n\n", d.Name, strings.Join(strs, "\n"), d.Name, d.Name)
}
