package main

import (
	"flag"
	"path/filepath"
	"strings"

	_ "framework/configure/pb"
	"framework/library/convertor"
	"framework/library/uerror"
	"framework/library/util"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

func main() {
	var src, dst string
	flag.StringVar(&src, "src", ".", "源目录")
	flag.StringVar(&dst, "dst", ".", "目的目录")
	flag.Parse()

	// 输出
	if ext := filepath.Ext(dst); len(ext) > 0 {
		dst = filepath.Dir(dst)
	}

	// 加载所有xlsx文件
	files, err := util.Glob(src, ".*\\.xlsx", false)
	if err != nil {
		panic(err)
	}

	// 解析文件
	p := NewMsgParser()
	for _, filename := range files {
		if err := p.ParseFile(filename); err != nil {
			panic(err)
		}
	}

	// 生成文件
	if err := p.Gen(dst); err != nil {
		panic(err)
	}
}

func GetAryMessageType(name string) (protoreflect.MessageType, error) {
	fullname := "bit_casino_golang." + name + "Ary"
	return protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(fullname))
}

func GetMessageType(name string) (protoreflect.MessageType, error) {
	fullname := "bit_casino_golang." + name
	return protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(fullname))
}

type Token int32

const (
	IDENT   Token = 0
	POINTER Token = 1
	ARRAY   Token = 2
	MAP     Token = 3
)

type Field struct {
	fieldType protoreflect.FieldDescriptor
	token     Token
	typename  string
	name      string
	position  int32
}

type StructDescriptor struct {
	aryType protoreflect.MessageType
	cfgType protoreflect.MessageType
	name    string
	list    []*Field
	data    map[string]*Field
	rows    [][]string
}

func NewStructDescriptor(name string, ary, cfg protoreflect.MessageType, rows [][]string) *StructDescriptor {
	return &StructDescriptor{
		aryType: ary,
		cfgType: cfg,
		name:    name,
		data:    make(map[string]*Field),
		rows:    rows,
	}
}

func (d *StructDescriptor) Put(pos int32, name, tname string) {
	item := d.parse(tname)
	item.name = name
	item.position = pos
	item.fieldType = d.cfgType.Descriptor().Fields().ByName(protoreflect.Name(name))
	d.list = append(d.list, item)
	d.data[item.name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{typename: item.typename, token: ARRAY}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{typename: item.typename, token: IDENT}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{typename: item.typename, token: POINTER}
	}
	return &Field{typename: convertor.Target(str), token: IDENT}
}

func (d *StructDescriptor) Marshal() ([]byte, error) {
	ary := dynamicpb.NewMessage(d.aryType.Descriptor())
	list := ary.Mutable(d.aryType.Descriptor().Fields().ByName("Ary")).List()
	for _, line := range d.rows {
		cfg := dynamicpb.NewMessage(d.cfgType.Descriptor())
		for _, field := range d.list {
			switch field.token {
			case IDENT, POINTER:
				cfg.Set(field.fieldType, convert(field, line[field.position-1]))
			case ARRAY:
				fieldList := cfg.Mutable(field.fieldType).List()
				for _, vv := range strings.Split(line[field.position-1], "|") {
					fieldList.Append(convert(field, vv))
				}
			}
		}
		list.Append(protoreflect.ValueOf(cfg))
	}

	marshaler := prototext.MarshalOptions{Multiline: true}
	return marshaler.Marshal(ary)
}

func convert(field *Field, val string) protoreflect.Value {
	value := convertor.Convert(field.typename, val)
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

type Value struct {
	typename string
	name     string
	value    int32
	desc     string
}

type EnumDescriptor struct {
	name string
	list []*Value
	data map[string]*Value
}

func NewEnumDescriptor(name string) *EnumDescriptor {
	return &EnumDescriptor{
		name: name,
		data: make(map[string]*Value),
	}
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *EnumDescriptor) Put(val int32, name string, gameType string, desc string) {
	item := &Value{
		typename: gameType,
		name:     name,
		value:    val,
		desc:     desc,
	}
	d.list = append(d.list, item)
	d.data[item.desc] = item
}

func (d *EnumDescriptor) ToInt32(val string) int32 {
	if val, ok := d.data[val]; ok {
		return val.value
	}
	return 0
}

type MsgParser struct {
	data    map[string]*EnumDescriptor
	enums   []*EnumDescriptor
	configs []*StructDescriptor
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]*EnumDescriptor),
	}
}

// @config[:col]|sheet:MessageName
// @enum|sheet
func (d *MsgParser) ParseFile(filename string) error {
	fp, err := excelize.OpenFile(filename)
	if err != nil {
		return uerror.Err(-1, "打开文件(%s)失败：%v", filename, err)
	}
	defer fp.Close()

	// 读取生成表
	rows, err := fp.GetRows("生成表")
	if err != nil {
		return uerror.Err(-1, "文件(%s)生成表不存在：%v", filename, err)
	}

	for _, items := range rows {
		for _, val := range items {
			if !strings.HasPrefix(val, "@") {
				continue
			}
			strs := strings.Split(val, "|")
			switch strings.ToLower(strs[0]) {
			case "@config":
				names := strings.Split(strs[1], ":")
				rows, err := fp.GetRows(names[0])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				if err := d.parseStruct(names[1], rows); err != nil {
					return err
				}
			case "@config:col":
				names := strings.Split(strs[1], ":")
				rows, err := fp.GetCols(names[0])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				if err := d.parseStruct(names[1], rows); err != nil {
					return err
				}
			case "@enum":
				rows, err := fp.GetRows(strs[1])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				d.parseEnum(rows)
			}
		}
	}
	return nil
}

func (d *MsgParser) parseStruct(name string, rows [][]string) error {
	aryType, err := GetAryMessageType(name)
	if err != nil {
		return err
	}
	cfgType, err := GetMessageType(name)
	if err != nil {
		return err
	}
	st := &StructDescriptor{
		aryType: aryType,
		cfgType: cfgType,
		name:    name,
		data:    make(map[string]*Field),
		rows:    rows[3:],
	}
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(int32(i)+1, rows[0][i], item)
	}
	d.configs = append(d.configs, st)
	return nil
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *MsgParser) parseEnum(rows [][]string) {
	for _, items := range rows {
		for _, val := range items {
			if !strings.HasPrefix(val, "E|") && !strings.HasPrefix(val, "e|") {
				continue
			}
			strs := strings.Split(val, "|")
			enum, ok := d.data[strs[2]]
			if !ok {
				enum = NewEnumDescriptor(strs[2])
				convertor.Register(func(val string) any { return enum.ToInt32(val) }, "int32", strs[2])
				d.enums = append(d.enums, enum)
				d.data[strs[2]] = enum
			}
			enum.Put(cast.ToInt32(strs[4]), strs[3], strs[2], strs[1])
		}
	}
}

func (d *MsgParser) Gen(dst string) error {
	for _, st := range d.configs {
		buf, err := st.Marshal()
		if err != nil {
			return err
		}
		if err := util.Save(dst, st.name+".conf", buf); err != nil {
			return err
		}
	}
	return nil
}
