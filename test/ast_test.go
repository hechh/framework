package test

import (
	"context"
	"fmt"
	"framework/library/descriptor"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type Parser struct{}

func (d *Parser) Visit(n ast.Node) ast.Visitor {
	switch vv := n.(type) {
	case *ast.File:
		return d
	case *ast.GenDecl:
		return d
	case *ast.TypeSpec:
		desc := descriptor.ParseType(vv.Type)
		fmt.Println(vv.Name.Name, "------", desc.Kind())
		/*
			for _, item := range desc.Members() {
				fmt.Println(item.Name, "======>", item.Type.Name())
			}
		*/

		switch vv.Type.(type) {
		case *ast.StructType:
		}
		return nil
	}
	return nil
}

func TestParser(t *testing.T) {
	fset := token.NewFileSet()
	fs, err := parser.ParseFile(fset, "../packet/packet.pb.go", nil, parser.ParseComments)
	if err != nil {
		t.Logf("错误：%v", err)
		return
	}

	// 遍历语法树
	ast.Walk(&Parser{}, fs)
}

func TestCompiler(t *testing.T) {
	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: []string{
				"../packet",
			},
		},
	}
	files, err := compiler.Compile(context.Background(), "packet.proto", "../test/client.proto")
	if err != nil {
		t.Logf("错误：%v", err)
		return
	}

	msgs := files[1].Messages()
	desc := msgs.ByName(protoreflect.Name("LoginReq"))
	obj := dynamicpb.NewMessage(desc)
	nameField := desc.Fields().ByName("Name")
	obj.Set(nameField, protoreflect.ValueOf("ttt"))

	buf, err := protojson.Marshal(obj)
	t.Log(string(buf), err)
}
