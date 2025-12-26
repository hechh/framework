package typespec

type Value struct {
	Type  string
	Name  string
	Value int32
	Desc  string
}

type EnumDescriptor struct {
	Name string
	list []*Value
	data map[string]*Value
}

func NewEnumDescriptor(Name string) *EnumDescriptor {
	return &EnumDescriptor{
		Name: Name,
		data: make(map[string]*Value),
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
	d.list = append(d.list, item)
	d.data[item.Desc] = item
}

func (d *EnumDescriptor) ToInt32(val string) int32 {
	if val, ok := d.data[val]; ok {
		return val.Value
	}
	return 0
}
