package typespec

import (
	"fmt"
	"framework/tools/cfgtool/domain"
	"framework/tools/cfgtool/internal/convert"
	"sort"
	"strings"
)

type MsgField struct {
	Typename string
	Name     string
	Position int32
	Desc     string
}

type StructMsgDescriptor struct {
	Type domain.Kind
	Name string
	List []*MsgField
	Data map[string]*MsgField
}

func NewStructMsgDescriptor(sheet string, Name string) *StructMsgDescriptor {
	return &StructMsgDescriptor{
		Type: domain.STRUCT,
		Name: Name,
		Data: make(map[string]*MsgField),
	}
}

func (d *StructMsgDescriptor) Kind() domain.Kind {
	return d.Type
}

func (d *StructMsgDescriptor) Put(pos int32, Name, tname, desc string) {
	item := d.parse(tname)
	item.Name = Name
	item.Position = pos
	item.Desc = desc
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructMsgDescriptor) parse(str string) *MsgField {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &MsgField{Typename: "repeated " + item.Typename}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &MsgField{Typename: item.Typename}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &MsgField{Typename: item.Typename}
	}
	return &MsgField{Typename: convert.Target(str)}
}

func (d *StructMsgDescriptor) String() string {
	sort.Slice(d.List, func(i int, j int) bool {
		return d.List[i].Position < d.List[j].Position
	})
	strs := []string{}
	for _, item := range d.List {
		strs = append(strs, fmt.Sprintf("\t%s %s = %d;\t// %s", item.Typename, item.Name, item.Position, item.Desc))
	}
	return fmt.Sprintf("message %s {\n%s\n}\n\nmessage %sAry {\nrepeated %s Ary = 1;\n}\n\n", d.Name, strings.Join(strs, "\n"), d.Name, d.Name)
}
