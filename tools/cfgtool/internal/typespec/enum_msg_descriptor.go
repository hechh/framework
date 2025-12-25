package typespec

import (
	"fmt"
	"framework/tools/cfgtool/domain"
	"sort"
	"strings"
)

type MsgValue struct {
	Typename string
	Name     string
	Value    int32
	Desc     string
}

type EnumMsgDescriptor struct {
	Type domain.Kind
	Name string
	List []*MsgValue
	Data map[string]*MsgValue
}

func NewEnumMsgDescriptor(name string) *EnumMsgDescriptor {
	return &EnumMsgDescriptor{
		Type: domain.ENUM,
		Name: name,
		Data: make(map[string]*MsgValue),
	}
}

func (d *EnumMsgDescriptor) Kind() domain.Kind {
	return d.Type
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *EnumMsgDescriptor) Put(val int32, name string, gameType string, desc string) {
	item := &MsgValue{
		Typename: gameType,
		Name:     name,
		Value:    val,
		Desc:     desc,
	}
	d.List = append(d.List, item)
	d.Data[item.Desc] = item
}

func (d *EnumMsgDescriptor) String() string {
	sort.Slice(d.List, func(i int, j int) bool {
		return d.List[i].Value < d.List[j].Value
	})
	strs := []string{}
	for _, item := range d.List {
		strs = append(strs, fmt.Sprintf("\t%s\t=\t%d;\t// %s", item.Name, item.Value, item.Desc))
	}
	return fmt.Sprintf("enum %s {\n%s\n}\n\n", d.Name, strings.Join(strs, "\n"))
}
