package main

import (
	"bytes"
	"flag"
	"framework/library/util"
	"path/filepath"
)

const (
	header = `
/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

syntax = "proto3";

package bit_casino_golang;

option  go_package = "./pb";

`
)

var (
	converts = map[string]string{
		"timestamp": "int64",
	}
)

func Convert(str string) string {
	if val, ok := converts[str]; ok {
		return val
	}
	return str
}

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
	parse := NewParseDescriptor()
	for _, filename := range files {
		if err := parse.ParseFile(filename); err != nil {
			panic(err)
		}
	}

	// 生成enum.gen.proto文件
	buf := bytes.NewBuffer(nil)
	buf.WriteString(header)
	if err := parse.GenEnum(buf, dst, "enum.gen.proto"); err != nil {
		panic(err)
	}

	// 生成table.gen.proto文件
	buf.Reset()
	buf.WriteString(header)
	if err := parse.GenTable(buf, dst, "enum.gen.proto"); err != nil {
		panic(err)
	}
}
