package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "framework/configure/pb"
	"framework/library/convertor"
	"framework/library/uerror"
	"framework/library/util"
	"html/template"
	"path"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	ProtoPkg = "bit_casino_golang"
)

type Kind int32

const (
	MAP   Kind = 1
	GROUP Kind = 2
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

	tpl := template.Must(template.New("CodeTpl").Funcs(template.FuncMap{
		"ToSnake":      strcase.ToSnake,
		"ToLowerCamel": strcase.ToLowerCamel,
	}).Parse(codeTpl))

	// 生成文件
	if err := p.Gen(dst, tpl); err != nil {
		panic(err)
	}
}

type MsgParser struct {
	data map[string]*StructDescriptor
	list []*StructDescriptor
}

func NewMsgParser() *MsgParser {
	return &MsgParser{
		data: make(map[string]*StructDescriptor),
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
	cfgType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(ProtoPkg + "." + name))
	if err != nil {
		return err
	}
	st := NewStructDescriptor(name, cfgType)
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

type Field struct {
	protoreflect.FieldDescriptor
	Name string
	Type string
}

type Index struct {
	Kind Kind
	Name string
	Type string
	List []*Field
}

type StructDescriptor struct {
	protoreflect.MessageType
	Name string
	Data map[string]*Index
	List []*Index
	data map[string]*Field
	list []*Field
}

func NewStructDescriptor(name string, msgType protoreflect.MessageType) *StructDescriptor {
	return &StructDescriptor{
		MessageType: msgType,
		Name:        name,
		Data:        make(map[string]*Index),
		data:        make(map[string]*Field),
	}
}

func (d *StructDescriptor) AddIndex(rule string) {
	strs := strings.Split(rule, ":")
	item := &Index{Name: strings.ReplaceAll(strs[1], ",", "")}
	for _, name := range strings.Split(strs[1], ",") {
		item.List = append(item.List, d.data[name])
	}
	switch strings.ToLower(strs[0]) {
	case "map":
		item.Kind = MAP
		item.Type = fmt.Sprintf("structure.Map%d", len(item.List))
	case "group":
		item.Kind = GROUP
		item.Type = fmt.Sprintf("structure.Group%d", len(item.List))
	}
	d.List = append(d.List, item)
	d.Data[item.Name] = item
}

func (d *StructDescriptor) Put(pos int32, name, tname string) {
	item := d.parse(tname)
	item.Name = name
	item.FieldDescriptor = d.Descriptor().Fields().ByName(protoreflect.Name(name))
	d.list = append(d.list, item)
	d.data[item.Name] = item
}

func (d *StructDescriptor) parse(str string) *Field {
	if strings.HasPrefix(str, "[]") {
		item := d.parse(strings.TrimPrefix(str, "[]"))
		return &Field{Type: item.Type}
	}
	if strings.HasPrefix(str, "&") {
		item := d.parse(strings.TrimPrefix(str, "&"))
		return &Field{Type: item.Type}
	}
	if strings.HasPrefix(str, "*") {
		item := d.parse(strings.TrimPrefix(str, "*"))
		return &Field{Type: item.Type}
	}
	return &Field{Type: convertor.Target(str)}
}

func (d *Index) GetArg() string {
	strs := []string{}
	for _, val := range d.List {
		if val.Kind() == protoreflect.EnumKind {
			strs = append(strs, val.Name+" pb."+val.Type)
		} else {
			strs = append(strs, val.Name+" "+val.Type)
		}
	}
	return strings.Join(strs, ", ")
}

func (d *Index) GetType() string {
	strs := []string{}
	for _, val := range d.List {
		if val.Kind() == protoreflect.EnumKind {
			strs = append(strs, "pb."+val.Type)
		} else {
			strs = append(strs, val.Type)
		}
	}
	return strings.Join(strs, ", ")
}

func (d *Index) GetValue(ref string) string {
	strs := []string{}
	for _, val := range d.List {
		if len(ref) > 0 {
			strs = append(strs, ref+"."+val.Name)
		} else {
			strs = append(strs, val.Name)
		}
	}
	return strings.Join(strs, ", ")
}

const codeTpl = `
/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

{{$type := .Name}}

package {{ToSnake $type}}

import (
	"framework/configure/pb"
	"framework/library/structure"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
)

var obj = atomic.Pointer[{{$type}}Data]{}

type {{$type}}Data struct {
	list []*pb.{{$type}}
{{- range $index := .List}}
	{{ToLowerCamel $index.Name}} {{$index.Type}}[{{$index.GetType}}, *pb.{{$type}}]
{{- end}}
}

func DeepCopy(item *pb.{{$type}}) *pb.{{$type}} {
	buf, _ := proto.Marshal(item)
	ret := &pb.{{$type}}{}
	proto.Unmarshal(buf, ret)
	return ret
}

func parse(buf []byte) error {
	ary := &pb.{{$type}}Ary{}
	if err := proto.UnmarshalText(string(buf), ary); err != nil {
		return err	
	}

	data := &{{$type}}Data{
{{- range $index := .List}}
		{{ToLowerCamel $index.Name}}: make({{$index.Type}}[{{$index.GetType}}, *pb.{{$type}}]),
{{- end}}
	}
	for _, item := range ary.Ary {
		data.list = append(data.list, item)
	{{- range $index := .List}}
		data.{{ToLowerCamel $index.Name}}.Put({{$index.GetValue "item"}}, item)
	{{- end}}
	}
	obj.Store(data)
	return nil
}

func LGet() (rets []*pb.{{$type}}) {
	list := obj.Load().list
	rets = make([]*pb.{{$type}}, len(list))
	copy(rets, list)
	return
}

func Walk(f func(*pb.{{$type}})bool) {
	for _, item := range obj.Load().list {
		if !f(item) {
			return
		}	
	}
}

{{range $index := .List}}
{{if eq $index.Kind 1}}		{{/* map类型 */}}
func MGet{{$index.Name}}({{$index.GetArg}}) *pb.{{$type}} {
	data := obj.Load().{{ToLowerCamel $index.Name}}
	if value, ok := data.Get({{$index.GetValue ""}}); ok {
		return value
	}
	return nil
}

{{else if eq $index.Kind 2}}	{{/* group类型 */}}
func GGet{{$index.Name}}({{$index.GetArg}}) []*pb.{{$type}} {
	data := obj.Load().{{ToLowerCamel $index.Name}}
	if value, ok := data.Get({{$index.GetValue ""}}); ok {
		return value
	}
	return nil
}

func GWalk{{$index.Name}}({{$index.GetArg}}, f func(*pb.{{$type}})bool) {
	data := obj.Load().{{ToLowerCamel $index.Name}}
	if values, ok := data.Get({{$index.GetValue ""}}); ok {
		for _, item := range values {
			if !f(item) {
				return	
			}	
		}
	}
}
{{end}}
{{end}}

`
