package main

import (
	"flag"
	_ "framework/configure/pb"
	"framework/library/util"
	"framework/tools/xlsx_to_code/internal/parser"
	"framework/tools/xlsx_to_code/internal/templ"
	"path/filepath"
	"text/template"

	"github.com/iancoleman/strcase"
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
	p := parser.NewMsgParser()
	for _, filename := range files {
		if err := p.ParseFile(filename); err != nil {
			panic(err)
		}
	}

	tpl := template.Must(template.New("CodeTpl").Funcs(template.FuncMap{
		"ToSnake":      strcase.ToSnake,
		"ToLowerCamel": strcase.ToLowerCamel,
	}).Parse(templ.Templ))

	// 生成文件
	if err := p.Gen(dst, tpl); err != nil {
		panic(err)
	}
}
