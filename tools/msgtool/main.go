package main

import (
	"bytes"
	"flag"
	"fmt"
	"framework/library/convertor"
	"framework/library/uerror"
	"framework/library/util"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
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
	files, err := util.Glob(src, ".*\\.xlsx", true)
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

// Field 结构体表示Proto文件中的一个字段定义
type Field struct {
	typename string
	name     string
	position int32
	desc     string
}

// StructDescriptor 结构体表示Proto文件中的一个message定义
type StructDescriptor struct {
	name string
	list []*Field
	data map[string]*Field
}

func NewStructDescriptor(Name string) *StructDescriptor {
	return &StructDescriptor{
		name: Name,
		data: make(map[string]*Field),
	}
}

func (d *StructDescriptor) Put(pos int32, Name, tname, desc string) {
	item := d.parse(tname)
	item.name = Name
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
	return &Field{typename: convertor.Target(str)}
}

func (d *StructDescriptor) String() string {
	sort.Slice(d.list, func(i int, j int) bool {
		return d.list[i].position < d.list[j].position
	})
	strs := []string{}
	for _, item := range d.list {
		strs = append(strs, fmt.Sprintf("\t%s %s = %d;\t// %s",
			item.typename, item.name, item.position, item.desc))
	}
	return fmt.Sprintf("message %s {\n%s\n}\n\nmessage %sAry {\nrepeated %s Ary = 1;\n}\n\n", d.name, strings.Join(strs, "\n"), d.name, d.name)
}

type IDescriptor interface {
	Put(int32, string, string, string)
	String() string
}

type MsgParser struct {
	data    map[string]IDescriptor
	enums   []*EnumDescriptor
	configs []*StructDescriptor
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]IDescriptor),
	}
}

func (d *MsgParser) ParseFile(filename string) error {
	// 打开文件
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

	// 遍历需要解析的sheet
	for _, items := range rows {
		for _, val := range items {
			if !strings.HasPrefix(val, "@") {
				continue
			}
			// @config[:col]|sheet:MessageName
			// @enum[:col]|sheet
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

func (d *MsgParser) parseStruct(name string, rows [][]string) {
	st := NewStructDescriptor(name)
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(int32(i)+1, rows[0][i], item, rows[2][i])
	}
	d.data[name] = st
	d.configs = append(d.configs, st)
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
				item := NewEnumDescriptor(strs[2])
				d.enums = append(d.enums, item)
				d.data[strs[2]] = item
				enum = item
			}
			enum.Put(cast.ToInt32(strs[4]), strs[3], strs[2], strs[1])
		}
	}
}

func (d *MsgParser) Gen(dst string) error {
	buf := bytes.NewBuffer(nil)
	pos, _ := buf.WriteString(`
/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

syntax = "proto3";

package bit_casino_golang;

option  go_package = "./pb";

	`)

	if len(d.enums) > 0 {
		sort.Slice(d.enums, func(i, j int) bool {
			return strings.Compare(d.enums[i].name, d.enums[j].name) <= 0
		})
		for _, item := range d.enums {
			buf.WriteString(item.String())
		}
		if err := util.Save(dst, "enum.gen.proto", buf.Bytes()); err != nil {
			return err
		}
	}

	if len(d.configs) > 0 {
		buf.Truncate(pos)
		if d.hasEnum() {
			buf.WriteString("import \"enum.gen.proto\";\n\n")
		}
		sort.Slice(d.configs, func(i, j int) bool {
			return strings.Compare(d.configs[i].name, d.configs[j].name) <= 0
		})
		for _, item := range d.configs {
			buf.WriteString(item.String())
		}
		return util.Save(dst, "table.gen.proto", buf.Bytes())
	}
	return nil
}

func (d *MsgParser) hasEnum() bool {
	for _, item := range d.configs {
		for _, field := range item.list {
			if aa, ok := d.data[field.typename]; ok {
				if _, ok := aa.(*EnumDescriptor); ok {
					return true
				}
			}
		}
	}
	return false
}
