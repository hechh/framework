package typespec

import (
	"fmt"
	"framework/tools/cfgtool/domain"
	"framework/tools/cfgtool/internal/convert"
	"sort"
	"strings"
)

type Field struct {
	Typename string
	Name     string
	Position int32
	Desc     string
}

type StructDescriptor struct {
	Type domain.Kind
	Name string
	List []*Field
	Data map[string]*Field
}

func NewStructDescriptor(sheet string, Name string) *StructDescriptor {
	return &StructDescriptor{
		Type: domain.STRUCT,
		Name: Name,
		Data: make(map[string]*Field),
	}
}

func (d *StructDescriptor) Kind() domain.Kind {
	return d.Type
}

func (d *StructDescriptor) Put(pos int32, Name, tname, desc string) {
	item := d.parse(tname)
	item.Name = Name
	item.Position = pos
	item.Desc = desc
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{Typename: "repeated " + item.Typename}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{Typename: item.Typename}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{Typename: item.Typename}
	}
	return &Field{Typename: convert.Target(str)}
}

func (d *StructDescriptor) String() string {
	sort.Slice(d.List, func(i int, j int) bool {
		return d.List[i].Position < d.List[j].Position
	})
	strs := []string{}
	for _, item := range d.List {
		strs = append(strs, fmt.Sprintf("\t%s %s = %d;\t// %s", item.Typename, item.Name, item.Position, item.Desc))
	}
	return fmt.Sprintf("message %s {\n%s\n}\n\nmessage %sAry {\nrepeated %s Ary = 1;\n}\n\n", d.Name, strings.Join(strs, "\n"), d.Name, d.Name)
}
