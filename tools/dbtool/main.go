package main

import (
	"bytes"
	"flag"
	"fmt"
	"framework/library/util"
	"framework/tools/dbtool/tpl"
	"go/ast"
	"html/template"
	"path"
	"strings"

	"github.com/iancoleman/strcase"
)

type Field struct {
	Name string
	Type string
}

// @dbtool:string|数据库|key:name@type|#房间id生成器
type RedisString struct {
	Pkg    string
	Name   string
	DbName string
	Format string
	Keys   []*Field
}

// @dbtool:hash|数据库|key:name@type|field:name@type|#房间id生成器
type RedisHash struct {
	Pkg      string
	Name     string
	DbName   string   // 数据库名字
	KeyFmt   string   // redis的key格式
	Keys     []*Field // 做成key格式的字段
	FieldFmt string   // redis的field格式
	Fields   []*Field // redis的field字段
}

type Parser struct {
	rules []string
	strs  []*RedisString
	hashs []*RedisHash
}

func (d *Parser) Visit(n ast.Node) ast.Visitor {
	switch vv := n.(type) {
	case *ast.File:
		return d
	case *ast.GenDecl:
		if vv.Doc != nil {
			d.rules = d.rules[:0]
			for _, str := range vv.Doc.List {
				rule := strings.TrimSpace(strings.TrimPrefix(str.Text, "//"))
				if strings.HasPrefix(rule, "@dbtool") {
					d.rules = append(d.rules, rule)
				}
			}
			if len(d.rules) > 0 {
				return d
			}
		}
		return nil
	case *ast.TypeSpec:
		switch vv.Type.(type) {
		case *ast.StructType:
			for _, rule := range d.rules {
				strs := strings.Split(rule, "|")
				switch strings.ToLower(strs[0]) {
				case "@dbtool:string":
					ff, fields := ParseField(strs[2])
					d.strs = append(d.strs, &RedisString{
						Pkg:    strcase.ToSnake(vv.Name.Name),
						Name:   vv.Name.Name,
						DbName: strs[1],
						Format: ff,
						Keys:   fields,
					})
				case "@dbtool:hash":
					kk, keys := ParseField(strs[2])
					ff, fields := ParseField(strs[3])
					d.hashs = append(d.hashs, &RedisHash{
						Pkg:      strcase.ToSnake(vv.Name.Name),
						Name:     vv.Name.Name,
						DbName:   strs[1],
						KeyFmt:   kk,
						Keys:     keys,
						FieldFmt: ff,
						Fields:   fields,
					})
				}
			}
		}
		return nil
	}
	return nil
}

func ParseField(str string) (key string, rets []*Field) {
	pos := strings.Index(str, ":")
	if pos >= 0 {
		key = str[:pos]
	}
	for _, val := range strings.Split(str[pos+1:], ",") {
		pos = strings.Index(val, "@")
		switch strings.ToLower(val[pos+1:]) {
		case "string":
			key += ":%s"
		default:
			key += ":%d"
		}
		rets = append(rets, &Field{
			Name: val[:pos],
			Type: val[pos+1:],
		})
	}
	return
}

func GetValues(args ...[]*Field) string {
	if len(args) <= 0 {
		return ""
	}
	values := make([]string, 0)
	for _, list := range args {
		for _, item := range list {
			values = append(values, item.Name)
		}
	}
	return strings.Join(values, ",")
}

func GetArgs(keys ...[]*Field) string {
	if len(keys) <= 0 {
		return ""
	}
	args := make([]string, 0)
	for _, list := range keys {
		for _, item := range list {
			args = append(args, fmt.Sprintf("%s %s", item.Name, item.Type))
		}
	}
	return strings.Join(args, ",")
}

func main() {
	var pbpath, output string
	flag.StringVar(&pbpath, "pb", "", ".pb.go文件目录")
	flag.StringVar(&output, "out", "", ".gen.go文件目录")
	flag.Parse()
	if len(pbpath) <= 0 {
		panic(".pb.go文件目录为空")
	}

	// 加载文件
	files, err := util.Glob(pbpath, ".*\\.pb\\.go", true)
	if err != nil {
		panic(err)
	}

	// 解析文件
	parser := &Parser{}
	if err := util.ParseFiles(parser, files...); err != nil {
		panic(err)
	}

	// 加载模板
	funcs := template.FuncMap{
		"GetValues": GetValues,
		"GetArgs":   GetArgs,
	}
	strTpl := template.Must(template.New("str").Funcs(funcs).Parse(tpl.RedisStringTpl))
	hashTpl := template.Must(template.New("hash").Funcs(funcs).Parse(tpl.RedisHashTpl))

	// 生成文件
	buf := bytes.NewBuffer(nil)
	for _, item := range parser.strs {
		buf.Reset()
		strTpl.Execute(buf, item)
		filename := path.Join(item.Pkg, item.Name+".gen.go")
		if err := util.SaveGo(output, filename, buf.Bytes()); err != nil {
			panic(err)
		}
	}
	for _, item := range parser.hashs {
		buf.Reset()
		hashTpl.Execute(buf, item)
		filename := path.Join(item.Pkg, item.Name+".gen.go")
		if err := util.SaveGo(output, filename, buf.Bytes()); err != nil {
			panic(err)
		}
	}
}
