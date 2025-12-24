package main

import (
	"bytes"
	"fmt"
	"framework/library/uerror"
	"framework/library/util"
	"sort"
	"strings"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

type Kind int32

const (
	BASIC   Kind = 0 // 基本类型 (int, string, bool)
	ENUM    Kind = 1 // 枚举类型
	STRUCT  Kind = 2 // 结构体类型 (struct{})
	POINTER Kind = 3 // 指针类型 (*T)
	ARRAY   Kind = 4 // 数组类型 ([N]T)
	MAP     Kind = 5 // map数据类型
)

type Value struct {
	name  string
	value int32
	desc  string
}

type EnumDescriptor struct {
	typeKind Kind
	name     string
	list     []*Value
	data     map[string]*Value
}

func NewEnumDescriptor(name string) *EnumDescriptor {
	return &EnumDescriptor{
		typeKind: ENUM,
		name:     name,
		data:     make(map[string]*Value),
	}
}

func (d *EnumDescriptor) Kind() Kind {
	return d.typeKind
}

func (d *EnumDescriptor) Name() string {
	return d.name
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *EnumDescriptor) Put(val int, name string, gameType string, desc string) {
	item := &Value{
		name:  name,
		value: int32(val),
		desc:  desc,
	}
	d.list = append(d.list, item)
	d.data[item.desc] = item
}

func (d *EnumDescriptor) String() string {
	sort.Slice(d.list, func(i int, j int) bool {
		return d.list[i].value < d.list[j].value
	})
	strs := []string{}
	for _, item := range d.list {
		strs = append(strs, fmt.Sprintf("\t%s\t=\t%d;\t// %s", item.name, item.value, item.desc))
	}
	return fmt.Sprintf("enum %s {\n%s\n}\n\n", d.name, strings.Join(strs, "\n"))
}

type Field struct {
	typename string
	name     string
	desc     string
	position int
}

type StructDescriptor struct {
	typeKind Kind
	name     string
	list     []*Field
	data     map[string]*Field
}

func NewStructDescriptor(name string) *StructDescriptor {
	return &StructDescriptor{
		typeKind: STRUCT,
		name:     name,
		data:     make(map[string]*Field),
	}
}

func (d *StructDescriptor) Kind() Kind {
	return d.typeKind
}

func (d *StructDescriptor) Name() string {
	return d.name
}

func (d *StructDescriptor) Put(pos int, name, tname, desc string) {
	item := d.parse(tname)
	item.name = name
	item.position = pos
	item.desc = desc
	d.list = append(d.list, item)
	d.data[item.name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{typename: "repeated " + item.typename}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{typename: item.typename}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{typename: item.typename}
	}
	return &Field{typename: Convert(str)}
}

func (d *StructDescriptor) String() string {
	sort.Slice(d.list, func(i int, j int) bool {
		return d.list[i].position < d.list[j].position
	})
	strs := []string{}
	for _, item := range d.list {
		strs = append(strs, fmt.Sprintf("\t%s %s = %d;\t// %s", item.typename, item.name, item.position, item.desc))
	}
	return fmt.Sprintf("message %s {\n%s\n}\n\nmessage %sAry {\nrepeated %s Ary = 1;\n}\n\n", d.name, strings.Join(strs, "\n"), d.name, d.name)
}

type IDescriptor interface {
	Kind() Kind
	Name() string
	Put(int, string, string, string)
	String() string
}

type ParseDescriptor struct {
	data  map[string]IDescriptor
	enums []*EnumDescriptor
	sts   []*StructDescriptor
}

func NewParseDescriptor() *ParseDescriptor {
	return &ParseDescriptor{data: make(map[string]IDescriptor)}
}

/*
@config[:col]|sheet:MessageName
@enum[:col]|sheet
*/
func (d *ParseDescriptor) ParseFile(filename string) error {
	fp, err := excelize.OpenFile(filename)
	if err != nil {
		return uerror.Wrap(-1, err)
	}
	defer fp.Close()

	// 读取需要生成proto文件地sheet
	rows, err := fp.GetRows("生成表")
	if err != nil {
		return uerror.Err(-1, "生成表不存在：%v")
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
				d.parseStruct(names[1], rows)
			case "@config:col":
				names := strings.Split(strs[1], ":")
				rows, err := fp.GetCols(names[0])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				d.parseStruct(names[1], rows)
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

func (d *ParseDescriptor) parseStruct(name string, rows [][]string) {
	st := NewStructDescriptor(name)
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(i+1, rows[0][i], item, rows[2][i])
	}
	d.data[name] = st
	d.sts = append(d.sts, st)
}

// E|游戏类型-德州NORMAL|GameType|Normal|1
func (d *ParseDescriptor) parseEnum(rows [][]string) {
	for _, items := range rows {
		for _, val := range items {
			if !strings.HasPrefix(val, "E|") && !strings.HasPrefix(val, "e|") {
				continue
			}
			strs := strings.Split(val, "|")

			enum, ok := d.data[strs[2]]
			if !ok {
				item := NewEnumDescriptor(strs[2])
				d.data[strs[2]] = item
				d.enums = append(d.enums, item)
				enum = item
			}
			enum.Put(cast.ToInt(strs[4]), strs[3], strs[2], strs[1])
		}
	}
}

func (d *ParseDescriptor) GenEnum(buf *bytes.Buffer, dst string, filename string) error {
	if len(d.enums) <= 0 {
		return nil
	}

	// 排序
	sort.Slice(d.enums, func(i, j int) bool {
		return strings.Compare(d.enums[i].Name(), d.enums[j].Name()) <= 0
	})

	for _, item := range d.enums {
		buf.WriteString(item.String())
	}
	return util.Save(dst, filename, buf.Bytes())
}

func (d *ParseDescriptor) GenTable(buf *bytes.Buffer, dst string, filename string) error {
	if len(d.enums) <= 0 {
		return nil
	}

	if d.finish() {
		buf.WriteString("import \"enum.gen.proto\";\n\n")
	}

	// 排序
	sort.Slice(d.sts, func(i, j int) bool {
		return strings.Compare(d.sts[i].Name(), d.sts[j].Name()) <= 0
	})

	for _, item := range d.sts {
		buf.WriteString(item.String())
	}
	return util.Save(dst, filename, buf.Bytes())
}

func (d *ParseDescriptor) finish() bool {
	for _, item := range d.sts {
		for _, f := range item.list {
			aa, ok := d.data[f.typename]
			if ok && aa.Kind() == ENUM {
				return true
			}
		}
	}
	return false
}
