package descriptor

import (
	"fmt"
	"go/ast"
	"strings"
)

// AstExprToString 将 ast.Expr 转换为表示类型的字符串
// 支持基本类型、指针、数组、切片、映射、通道、函数、结构体、接口等复杂类型
func AstExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	// 基本标识符（int, string, MyType 等）
	case *ast.Ident:
		return t.Name
	// 指针类型 *T
	case *ast.StarExpr:
		return "*" + AstExprToString(t.X)
	// 数组和切片类型
	case *ast.ArrayType:
		if t.Len == nil {
			// 切片类型 []T
			return "[]" + AstExprToString(t.Elt)
		}
		// 数组类型 [N]T
		switch lenExpr := t.Len.(type) {
		case *ast.BasicLit:
			return "[" + lenExpr.Value + "]" + AstExprToString(t.Elt)
		case *ast.Ident:
			return "[" + lenExpr.Name + "]" + AstExprToString(t.Elt)
		default:
			return "[?]" + AstExprToString(t.Elt) // 无法确定的数组长度
		}
	// 映射类型 map[K]V
	case *ast.MapType:
		return "map[" + AstExprToString(t.Key) + "]" + AstExprToString(t.Value)
	// 通道类型
	case *ast.ChanType:
		var prefix string
		switch t.Dir {
		case ast.SEND:
			prefix = "chan<- "
		case ast.RECV:
			prefix = "<-chan "
		default:
			prefix = "chan "
		}
		return prefix + AstExprToString(t.Value)

	// 函数类型
	case *ast.FuncType:
		return funcTypeToString(t)
	// 结构体类型
	case *ast.StructType:
		return structTypeToString(t)
	// 接口类型
	case *ast.InterfaceType:
		return interfaceTypeToString(t)
	// 选择器表达式 pkg.Type
	case *ast.SelectorExpr:
		return AstExprToString(t.X) + "." + t.Sel.Name
	// 括号表达式 (T)
	case *ast.ParenExpr:
		return "(" + AstExprToString(t.X) + ")"
	// 不定参数 ...T
	case *ast.Ellipsis:
		return "..." + AstExprToString(t.Elt)
	// 泛型实例化类型 List[T]
	case *ast.IndexExpr:
		return AstExprToString(t.X) + "[" + AstExprToString(t.Index) + "]"
	// 多类型参数的泛型实例化 Map[K, V]
	case *ast.IndexListExpr:
		indices := make([]string, len(t.Indices))
		for i, index := range t.Indices {
			indices[i] = AstExprToString(index)
		}
		return AstExprToString(t.X) + "[" + strings.Join(indices, ", ") + "]"
	// 基本字面量（在类型上下文中可能遇到）
	case *ast.BasicLit:
		return t.Value
	default:
		// 对于未处理的情况，返回类型信息便于调试
		return fmt.Sprintf("/* %T */", expr)
	}
}

// funcTypeToString 将函数类型转换为字符串表示
func funcTypeToString(funcType *ast.FuncType) string {
	var buf strings.Builder
	buf.WriteString("func")

	// 参数列表
	buf.WriteString("(")
	if funcType.Params != nil {
		for i, field := range funcType.Params.List {
			if i > 0 {
				buf.WriteString(", ")
			}
			if len(field.Names) > 0 {
				// 有参数名
				names := make([]string, len(field.Names))
				for j, name := range field.Names {
					names[j] = name.Name
				}
				buf.WriteString(strings.Join(names, ", "))
				buf.WriteString(" ")
			}
			buf.WriteString(AstExprToString(field.Type))
		}
	}
	buf.WriteString(")")

	// 返回值
	if funcType.Results != nil {
		if len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
			// 单返回值，无名称
			buf.WriteString(" ")
			buf.WriteString(AstExprToString(funcType.Results.List[0].Type))
		} else if len(funcType.Results.List) > 0 {
			buf.WriteString(" (")
			for i, field := range funcType.Results.List {
				if i > 0 {
					buf.WriteString(", ")
				}
				if len(field.Names) > 0 {
					// 有返回值名
					names := make([]string, len(field.Names))
					for j, name := range field.Names {
						names[j] = name.Name
					}
					buf.WriteString(strings.Join(names, ", "))
					buf.WriteString(" ")
				}
				buf.WriteString(AstExprToString(field.Type))
			}
			buf.WriteString(")")
		}
	}

	return buf.String()
}

// structTypeToString 将结构体类型转换为字符串表示
func structTypeToString(structType *ast.StructType) string {
	if structType.Fields == nil || len(structType.Fields.List) == 0 {
		return "struct{}"
	}

	var buf strings.Builder
	buf.WriteString("struct {\n")

	for _, field := range structType.Fields.List {
		// 字段名
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				buf.WriteString("\t")
				buf.WriteString(name.Name)
				buf.WriteString(" ")
				buf.WriteString(AstExprToString(field.Type))
				if field.Tag != nil {
					buf.WriteString(" ")
					buf.WriteString(field.Tag.Value)
				}
				buf.WriteString("\n")
			}
		} else {
			// 嵌入字段
			buf.WriteString("\t")
			buf.WriteString(AstExprToString(field.Type))
			if field.Tag != nil {
				buf.WriteString(" ")
				buf.WriteString(field.Tag.Value)
			}
			buf.WriteString("\n")
		}
	}

	buf.WriteString("}")
	return buf.String()
}

// interfaceTypeToString 将接口类型转换为字符串表示
func interfaceTypeToString(interfaceType *ast.InterfaceType) string {
	if interfaceType.Methods == nil || len(interfaceType.Methods.List) == 0 {
		return "interface{}"
	}

	var buf strings.Builder
	buf.WriteString("interface {\n")

	for _, method := range interfaceType.Methods.List {
		buf.WriteString("\t")
		if len(method.Names) > 0 {
			// 方法
			buf.WriteString(method.Names[0].Name)
			buf.WriteString(AstExprToString(method.Type))
		} else {
			// 嵌入接口
			buf.WriteString(AstExprToString(method.Type))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("}")
	return buf.String()
}
