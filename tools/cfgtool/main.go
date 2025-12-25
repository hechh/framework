package main

import (
	"flag"
	"framework/library/util"
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
	p := parse.NewMsgParser()
	for _, filename := range files {
		if err := p.ParseFile(filename); err != nil {
			panic(err)
		}
	}

	// 解析完成
	p.Complete()

	// 生成文件
	if err := p.Gen(dst); err != nil {
		panic(err)
	}
}
