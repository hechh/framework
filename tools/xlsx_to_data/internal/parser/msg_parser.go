package parser

import (
	"framework/library/convertor"
	"framework/library/uerror"
	"framework/library/util"
	"framework/tools/xlsx_to_data/domain"
	"framework/tools/xlsx_to_data/internal/typespec"
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
	return &MsgParser{data: make(map[string]*typespec.EnumDescriptor)}
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
	aryType, err := domain.GetMessageType(name + "Ary")
	if err != nil {
		return err
	}
	cfgType, err := domain.GetMessageType(name)
	if err != nil {
		return err
	}
	st := typespec.NewStructDescriptor(name, aryType, cfgType, rows[3:])
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
				enum = typespec.NewEnumDescriptor(strs[2])
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
		if err := util.Save(dst, st.Name+".conf", buf); err != nil {
			return err
		}
	}
	return nil
}
