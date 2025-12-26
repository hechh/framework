package typespec

import (
	"framework/library/convertor"
	"framework/tools/xlsx_to_data/domain"
	"strings"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Field struct {
	fieldType protoreflect.FieldDescriptor
	Token     domain.Token
	Type      string
	Name      string
	Position  int32
}

type StructDescriptor struct {
	aryType protoreflect.MessageType
	cfgType protoreflect.MessageType
	Name    string
	List    []*Field
	Data    map[string]*Field
	rows    [][]string
}

func NewStructDescriptor(Name string, ary, cfg protoreflect.MessageType, rows [][]string) *StructDescriptor {
	return &StructDescriptor{
		aryType: ary,
		cfgType: cfg,
		Name:    Name,
		Data:    make(map[string]*Field),
		rows:    rows,
	}
}

func (d *StructDescriptor) Put(pos int32, Name, tname string) {
	item := d.parse(tname)
	item.Name = Name
	item.Position = pos
	item.fieldType = d.cfgType.Descriptor().Fields().ByName(protoreflect.Name(Name))
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{Type: item.Type, Token: domain.ARRAY}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{Type: item.Type, Token: domain.IDENT}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{Type: item.Type, Token: domain.POINTER}
	}
	return &Field{Type: convertor.Target(str), Token: domain.IDENT}
}

func (d *StructDescriptor) Marshal() ([]byte, error) {
	ary := dynamicpb.NewMessage(d.aryType.Descriptor())
	List := ary.Mutable(d.aryType.Descriptor().Fields().ByName("Ary")).List()
	for _, line := range d.rows {
		cfg := dynamicpb.NewMessage(d.cfgType.Descriptor())
		for _, field := range d.List {
			switch field.Token {
			case domain.IDENT, domain.POINTER:
				cfg.Set(field.fieldType, convert(field, line[field.Position-1]))
			case domain.ARRAY:
				fieldList := cfg.Mutable(field.fieldType).List()
				for _, vv := range strings.Split(line[field.Position-1], "|") {
					fieldList.Append(convert(field, vv))
				}
			}
		}
		List.Append(protoreflect.ValueOf(cfg))
	}

	marshaler := prototext.MarshalOptions{Multiline: true}
	return marshaler.Marshal(ary)
}

func convert(field *Field, val string) protoreflect.Value {
	value := convertor.Convert(field.Type, val)
	switch field.fieldType.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOf(value)
	case protoreflect.EnumKind:
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(value.(int32)))
	case protoreflect.Int32Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Sint32Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Uint32Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Int64Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Sint64Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Uint64Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Sfixed32Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Fixed32Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.FloatKind:
		return protoreflect.ValueOf(value)
	case protoreflect.Sfixed64Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.Fixed64Kind:
		return protoreflect.ValueOf(value)
	case protoreflect.DoubleKind:
		return protoreflect.ValueOf(value)
	case protoreflect.StringKind:
		return protoreflect.ValueOf(value)
	case protoreflect.BytesKind:
		return protoreflect.ValueOf(value)
	case protoreflect.MessageKind:
		return protoreflect.ValueOf(value)
	case protoreflect.GroupKind:
		return protoreflect.ValueOf(value)
	}
	return protoreflect.ValueOf(nil)
}
