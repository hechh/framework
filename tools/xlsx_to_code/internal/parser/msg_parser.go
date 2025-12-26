package parser

import (
	"bytes"
	"framework/library/uerror"
	"framework/library/util"
	"framework/tools/cfgtool/domain"
	"framework/tools/xlsx_to_code/internal/typespec"
	"path"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/xuri/excelize/v2"
)

type MsgParser struct {
	data map[string]*typespec.StructDescriptor
	list []*typespec.StructDescriptor
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]*typespec.StructDescriptor),
	}
}

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
				if err := d.parseStruct(names[1], rows, strs[2:]...); err != nil {
					return err
				}
			case "@config:col":
				names := strings.Split(strs[1], ":")
				rows, err := fp.GetCols(names[0])
				if err != nil {
					return uerror.Err(-1, "表格(%s)不存在", strs[1])
				}
				if err := d.parseStruct(names[1], rows, strs[2:]...); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *MsgParser) parseStruct(name string, rows [][]string, rules ...string) error {
	cfgType, err := domain.GetMessageType(name)
	if err != nil {
		return err
	}
	st := typespec.NewStructDescriptor(name, cfgType)
	for i, item := range rows[1] {
		if len(item) <= 0 {
			continue
		}
		st.Put(int32(i)+1, rows[0][i], item)
	}
	for _, rule := range rules {
		st.AddIndex(rule)
	}
	d.data[st.Name] = st
	d.list = append(d.list, st)
	return nil
}

func (d *MsgParser) Gen(dst string, tpl *template.Template) error {
	buf := bytes.NewBuffer(nil)
	for _, item := range d.list {
		pkgname := strcase.ToSnake(item.Name)
		if err := tpl.Execute(buf, item); err != nil {
			return err
		}

		if err := util.SaveGo(path.Join(dst, pkgname), item.Name+".gen.go", buf.Bytes()); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}
