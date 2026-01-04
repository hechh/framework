package parser

import (
	"bytes"
	"fmt"
	"framework/library/uerror"
	"framework/library/util"
	"framework/tools/xlsx_to_proto/domain"
	"framework/tools/xlsx_to_proto/internal/typespec"
	"sort"
	"strings"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

type MsgParser struct {
	data    map[string]*typespec.EnumDescriptor
	enums   []*typespec.EnumDescriptor
	configs []*typespec.StructDescriptor
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]*typespec.EnumDescriptor),
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
	st := typespec.NewStructDescriptor(name)
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(int32(i)+1, rows[0][i], item, rows[2][i])
	}
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
				enum = typespec.NewEnumDescriptor(strs[2])
				d.enums = append(d.enums, enum)
				d.data[strs[2]] = enum
			}
			enum.Put(cast.ToInt32(strs[4]), strs[3], strs[2], strs[1])
		}
	}
}

func (d *MsgParser) Gen(dst string) error {
	buf := bytes.NewBuffer(nil)
	pos, _ := buf.WriteString(fmt.Sprintf(header, domain.PROTO_PKG_NAME))

	if len(d.enums) > 0 {
		sort.Slice(d.enums, func(i, j int) bool {
			return strings.Compare(d.enums[i].Name, d.enums[j].Name) <= 0
		})
		for _, item := range d.enums {
			buf.WriteString(item.String())
		}
		if err := util.Save(dst, "enum.gen.proto", buf.Bytes()); err != nil {
			return err
		}
	}

	if len(d.configs) > 0 {
		sort.Slice(d.configs, func(i, j int) bool {
			return strings.Compare(d.configs[i].Name, d.configs[j].Name) <= 0
		})

		buf.Truncate(pos)
		if d.hasEnum() {
			buf.WriteString("import \"enum.gen.proto\";\n\n")
		}
		for _, item := range d.configs {
			buf.WriteString(item.String())
		}
		return util.Save(dst, "table.gen.proto", buf.Bytes())
	}
	return nil
}

func (d *MsgParser) hasEnum() bool {
	for _, item := range d.configs {
		for _, field := range item.List {
			if _, ok := d.data[field.Type]; ok {
				return true
			}
		}
	}
	return false
}

const (
	header = `
/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

syntax = "proto3";

package %s;
option  go_package = "./pb";

`
)
