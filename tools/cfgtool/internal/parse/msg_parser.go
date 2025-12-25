package parse

import (
	"bytes"
	"framework/library/uerror"
	"framework/library/util"
	"framework/tools/cfgtool/domain"
	"framework/tools/cfgtool/internal/typespec"
	"sort"
	"strings"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

type MsgParser struct {
	data    map[string]domain.IMsgDescriptor
	enums   []*typespec.EnumMsgDescriptor
	configs []*typespec.StructMsgDescriptor
	rows    map[string][][]string
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]domain.IMsgDescriptor),
		rows: make(map[string][][]string),
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
				d.parseStruct(names[0], names[1], rows)
			case "@config:col":
				names := strings.Split(strs[1], ":")
				rows, err := fp.GetCols(names[0])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				d.parseStruct(names[0], names[1], rows)
			case "@enum":
				rows, err := fp.GetRows(strs[1])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				d.parseEnum(strs[1], rows)
			}
		}
	}
	return nil
}

func (d *MsgParser) parseStruct(sheet, name string, rows [][]string) {
	st := typespec.NewStructMsgDescriptor(sheet, name)
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(int32(i)+1, rows[0][i], item, rows[2][i])
	}
	d.data[name] = st
	d.configs = append(d.configs, st)
	d.rows[name] = rows[3:]
}

func (d *MsgParser) parseEnum(sheet string, rows [][]string) {
	for _, items := range rows {
		for _, val := range items {
			if !strings.HasPrefix(val, "E|") && !strings.HasPrefix(val, "e|") {
				continue
			}
			// E|游戏类型-德州NORMAL|GameType|Normal|1
			strs := strings.Split(val, "|")
			enum, ok := d.data[strs[2]]
			if !ok {
				item := typespec.NewEnumMsgDescriptor(strs[2])
				d.enums = append(d.enums, item)
				d.data[strs[2]] = item
				enum = item
			}
			enum.Put(cast.ToInt32(strs[4]), strs[3], strs[2], strs[1])
		}
	}
}

func (d *MsgParser) Complete() {
	// 排序
	sort.Slice(d.enums, func(i, j int) bool {
		return strings.Compare(d.enums[i].Name, d.enums[j].Name) <= 0
	})
	sort.Slice(d.configs, func(i, j int) bool {
		return strings.Compare(d.configs[i].Name, d.configs[j].Name) <= 0
	})
}

func (d *MsgParser) hasEnum() bool {
	for _, item := range d.configs {
		for _, field := range item.List {
			aa, ok := d.data[field.Typename]
			if ok && aa.Kind() == domain.ENUM {
				return true
			}
		}
	}
	return false
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
		for _, item := range d.configs {
			buf.WriteString(item.String())
		}
		return util.Save(dst, "table.gen.proto", buf.Bytes())
	}
	return nil
}

/*
func (d *MsgParser) GenData(buf *bytes.Buffer, dst string) error {
	for name, rows := range d.rows {
		aryType, cfgType, err := FindMessageByName(name)
		if err != nil {
			return err
		}
		if err := d.parse(aryType, cfgType, rows); err != nil {
			return err
		}
	}
	return nil
}

func (d *MsgParser) parse(aryType, cfgType protoreflect.MessageType, item domain.IDescriptor, rows [][]string) error {
	ary := dynamicpb.NewMessage(aryType.Descriptor())
	fields := item.Members()
	for _, line := range rows {
		cfg := dynamicpb.NewMessage(cfgType.Descriptor())
		for _, field := range fields {
			item := cfgType.Descriptor().Fields().ByName(protoreflect.Name(field.TypeName()))
			if item.Kind() == protoreflect.Int32Kind {
				cfg.Set(item, protoreflect.ValueOf(convert.Convert(field.TypeName(), line[field.Value()])))
			}
		}
	}

	return nil
}

func FindMessageByName(name string) (aryType, cfgType protoreflect.MessageType, err error) {
	aryfull := protoreflect.FullName(domain.ProtoPkgName + "Ary." + name)
	if aryType, err = protoregistry.GlobalTypes.FindMessageByName(aryfull); err != nil {
		return
	}
	cfgfull := protoreflect.FullName(domain.ProtoPkgName + "." + name)
	cfgType, err = protoregistry.GlobalTypes.FindMessageByName(cfgfull)
	return
}
*/
