package templ

const Templ = `
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

func SGet(pos int) *pb.{{$type}} {
	if pos < 0 {
		pos = 0	
	}
	list := obj.Load().list
	if ll := len(list); ll-1 < pos {
		pos = ll-1	
	}
	return list[pos]
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
