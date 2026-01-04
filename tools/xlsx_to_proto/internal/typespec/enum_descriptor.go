package typespec

import (
	"fmt"
	"sort"
	"strings"
)

type Value struct {
	Type  string
	Name  string
	Value int32
	Desc  string
}
type EnumDescriptor struct {
	Name string
	List []*Value
	Data map[string]*Value
}

func NewEnumDescriptor(Name string) *EnumDescriptor {
	return &EnumDescriptor{
		Name: Name,
		Data: make(map[string]*Value),
	}
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *EnumDescriptor) Put(val int32, Name string, gameType string, Desc string) {
	item := &Value{
		Type:  gameType,
		Name:  Name,
		Value: val,
		Desc:  Desc,
	}
	d.List = append(d.List, item)
	d.Data[item.Desc] = item
}

func (d *EnumDescriptor) String() string {
	sort.Slice(d.List, func(i int, j int) bool {
		return d.List[i].Value < d.List[j].Value
	})
	strs := []string{}
	for _, item := range d.List {
		strs = append(strs, fmt.Sprintf("\t%s\t=\t%d;\t// %s", item.Name, item.Value, item.Desc))
	}
	return fmt.Sprintf("enum %s {\n%s\n}\n\n", d.Name, strings.Join(strs, "\n"))
}
