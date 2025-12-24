package domain

type Kind int32

const (
	BASIC  Kind = 0 // 基本类型 (int, string, bool)
	ENUM   Kind = 1 // 枚举类型
	STRUCT Kind = 2 // 结构体类型 (struct{})
)

type Token int32

const (
	IDENT   Kind = 0
	POINTER Kind = 1 // 指针类型 (*T)
	ARRAY   Kind = 2 // 数组类型 ([N]T)
	MAP     Kind = 3 // map数据类型
)

type Descriptor interface {
	Kind() Kind
	Put(int32, string, string, string)
	String() string
}

const (
	ProtoPkgName = "bit_casino_golang"
	Header       = `
/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

syntax = "proto3";

package bit_casino_golang;

option  go_package = "./pb";

`
)
