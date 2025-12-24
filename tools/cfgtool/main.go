package main

import (
	"bytes"
	"flag"
	"framework/library/util"
	"framework/tools/cfgtool/domain"
	"framework/tools/cfgtool/internal/parse"
	"path/filepath"
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
	p := parse.NewParser()
	for _, filename := range files {
		if err := p.ParseFile(filename); err != nil {
			panic(err)
		}
	}

	// 生成enum.gen.proto文件
	buf := bytes.NewBuffer(nil)
	buf.WriteString(domain.Header)
	if err := p.GenEnum(buf, dst, "enum.gen.proto"); err != nil {
		panic(err)
	}

	// 生成table.gen.proto文件
	buf.Reset()
	buf.WriteString(domain.Header)
	if p.HasEnum() {
		buf.WriteString("import \"enum.gen.proto\";\n\n")
	}
	if err := p.GenTable(buf, dst, "table.gen.proto"); err != nil {
		panic(err)
	}
}
