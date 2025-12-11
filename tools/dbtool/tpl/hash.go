package tpl

const RedisHashTpl = `
/*
* 本代码由pbtool工具生成，请勿手动修改
*/

package {{.Pkg}}

import (
	"fmt"
	
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"
)

const (
	DBNAME = {{.DbName}}
)

func GetKey({{GetArgs .Keys}}) string {
	return fmt.Sprintf("{{.KeyFmt}}" {{if .Keys}},{{end}} {{GetValues .Keys}})
}

func GetField({{GetArgs .Fields}}) string {
	return fmt.Sprintf("{{.FieldFmt}}" {{if .Fields}},{{end}} {{GetValues .Fields}})
}



`
