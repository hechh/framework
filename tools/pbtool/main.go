package main

import (
	"bytes"
	"flag"
	"framework/library/descriptor"
	"framework/library/util"
	"go/ast"
	"strings"
	"text/template"
)

const tpl = `
/*
* 本代码由pbtool工具生成，请勿手动修改
*/

package {{.GetPkgName}}

import (
	"github.com/gogo/protobuf/proto"
)

{{range $st := .GetAllRsp -}}
{{range $field := $st.Members -}}
{{if eq $field.Type.Name "*RspHead" -}}
func (d *{{$st.Name}}) SetRspHead(v {{$field.Type.Name}}) {
	d.{{$field.Name}} = v
}
{{end}}
{{end}}
{{end}}

{{range $st := .GetAllStruct -}}
func(d *{{$st.Name}}) ToDB() ([]byte, error) {
	if d == nil {
		return nil, nil
	}
	return proto.Marshal(d)
}

func(d *{{$st.Name}}) FromDB(val []byte) error {
	if len(val) <= 0 {
		return nil
	}
	return proto.Unmarshal(val, d)
}

{{end}}
`

type StructDescriptor struct {
	Name string
	descriptor.TypeDescriptor
}

type Parser struct {
	pkgName string
	list    []*StructDescriptor
}

func (p *Parser) Visit(n ast.Node) ast.Visitor {
	switch vv := n.(type) {
	case *ast.File:
		p.pkgName = vv.Name.Name
		return p
	case *ast.GenDecl:
		return p
	case *ast.TypeSpec:
		switch vv.Type.(type) {
		case *ast.StructType:
			item := descriptor.ParseType(vv.Type)
			p.list = append(p.list, &StructDescriptor{
				Name:           vv.Name.Name,
				TypeDescriptor: item,
			})
		}
		return nil
	}
	return nil
}

func (p *Parser) GetPkgName() string {
	return p.pkgName
}

func (p *Parser) GetAllStruct() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if strings.HasSuffix(item.Name, "Rsp") || strings.HasSuffix(item.Name, "Req") || strings.HasSuffix(item.Name, "Config") || strings.HasSuffix(item.Name, "ConfigAry") {
			continue
		}
		rets = append(rets, item)
	}
	return
}

func (p *Parser) GetAllRsp() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if strings.HasSuffix(item.Name, "Rsp") {
			rets = append(rets, item)
		}
	}
	return
}

func main() {
	var pbpath string
	flag.StringVar(&pbpath, "pb", "", ".pb.go文件目录")
	flag.Parse()
	if len(pbpath) <= 0 {
		panic(".pb.go文件目录为空")
	}

	files, err := util.Glob(pbpath, ".*\\.pb\\.go", true)
	if err != nil {
		panic(err)
	}

	parser := &Parser{}
	if err := util.ParseFiles(parser, files...); err != nil {
		panic(err)
	}

	// 生成文件
	tplObj := template.Must(template.New("pb").Parse(tpl))
	buf := bytes.NewBuffer(nil)
	tplObj.Execute(buf, parser)
	if err := util.SaveGo(pbpath, "rsp_head.gen.pb.go", buf.Bytes()); err != nil {
		panic(err)
	}
}
