package domain

type Kind int32

const (
	BASIC   Kind = 0 // 基本类型 (int, string, bool)
	ENUM    Kind = 1 // 枚举类型
	STRUCT  Kind = 2 // 结构体类型 (struct{})
	POINTER Kind = 3 // 指针类型 (*T)
	ARRAY   Kind = 4 // 数组类型 ([N]T)
	MAP     Kind = 5 // map数据类型
)

type IMsgDescriptor interface {
	Kind() Kind
	Put(int32, string, string, string)
	String() string
}

/*
const (
	ProtoPkgName = "bit_casino_golang"
)
*/
