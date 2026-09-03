package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	futil "github.com/hechh/framework/library/fileutil"
)

type StructDescriptor struct {
	Name string
	TypeDescriptor
	// DBShard 结构体声明的 @pbtool dbpool 分库分表规则（无规则为 nil）
	DBShard *DBShard
}

// DBShard 描述 @pbtool dbpool 分库分表规则
// 注解格式: @pbtool:dbpool|shard:N|table:field@type
// 例: @pbtool:dbpool|shard:64|bingo_contest_result_data:uid@uint64
// 生成: TableName() 返回 fmt.Sprintf("bingo_contest_result_data_%d", d.Uid%64)
type DBShard struct {
	Table     string // 基础表名 (bingo_contest_result_data)
	Shard     int    // 分片数量 (64)
	ShardKey  string // 分片字段注解名 (uid)
	GoField   string // 分片字段 Go 导出名 (Uid)
	ShardType string // 分片字段类型 (uint64)
}

type EnumDescriptor struct {
	Name string
	TypeDescriptor
}

type Parser struct {
	pkgName string
	list    []*StructDescriptor
	enums   []*EnumDescriptor
	rules   []string // 当前 GenDecl 文档注释中的 @pbtool 规则（供紧随其后的 TypeSpec 使用）
}

func (p *Parser) Visit(n ast.Node) ast.Visitor {
	switch vv := n.(type) {
	case *ast.File:
		p.pkgName = vv.Name.Name
		return p
	case *ast.GenDecl:
		// 收集类型声明文档注释中的 @pbtool 注解
		p.rules = p.rules[:0]
		if vv.Doc != nil {
			for _, comment := range vv.Doc.List {
				line := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
				if strings.HasPrefix(line, "@pbtool") {
					p.rules = append(p.rules, line)
				}
			}
		}
		return p
	case *ast.TypeSpec:
		switch vv.Type.(type) {
		case *ast.StructType:
			item := ParseType(vv.Type)
			desc := &StructDescriptor{
				Name:           vv.Name.Name,
				TypeDescriptor: item,
			}
			for _, rule := range p.rules {
				if shard := parseDBShardRule(rule); shard != nil {
					if hasMember(item, shard.GoField) {
						desc.DBShard = shard
					} else {
						fmt.Printf("[pbtool] 忽略 %s 的 dbpool 规则: 结构体缺少分片字段 %s\n",
							desc.Name, shard.GoField)
					}
				}
			}
			p.list = append(p.list, desc)
		case *ast.Ident:
			item := ParseType(vv.Type)
			p.enums = append(p.enums, &EnumDescriptor{
				Name:           vv.Name.Name,
				TypeDescriptor: item,
			})
		}
		p.rules = p.rules[:0]
		return nil
	}
	return nil
}

// parseDBShardRule 解析 @pbtool:dbpool|shard:N|table:field@type 规则
// 返回 nil 表示非法的 dbpool 分库规则
func parseDBShardRule(rule string) *DBShard {
	body, found := strings.CutPrefix(rule, "@pbtool:")
	if !found {
		return nil
	}
	parts := strings.Split(body, "|")
	// 仅当首段是 dbpool 时才认定为本工具关心的规则，格式不完整时打印告警便于排查
	isDBPool := strings.TrimSpace(parts[0]) == "dbpool"
	if !isDBPool {
		return nil
	}
	if len(parts) != 3 {
		fmt.Printf("[pbtool] 忽略无效 dbpool 规则: %s (期望 3 段: dbpool|shard:N|table:field@type)\n", body)
		return nil
	}
	shardSpec, found := strings.CutPrefix(strings.TrimSpace(parts[1]), "shard:")
	if !found {
		fmt.Printf("[pbtool] 忽略无效 dbpool 规则: %s (第 2 段需为 shard:N)\n", body)
		return nil
	}
	shard, err := strconv.Atoi(strings.TrimSpace(shardSpec))
	if err != nil || shard <= 0 {
		fmt.Printf("[pbtool] 忽略无效 dbpool 分片数: %s\n", parts[1])
		return nil
	}
	// 第三段格式: table:field@type
	tableSpec := strings.TrimSpace(parts[2])
	sep := strings.LastIndex(tableSpec, ":")
	if sep <= 0 || sep >= len(tableSpec)-1 {
		fmt.Printf("[pbtool] 忽略无效 dbpool 表规则: %s\n", tableSpec)
		return nil
	}
	table := tableSpec[:sep]
	keySpec := tableSpec[sep+1:]
	field, fieldType := keySpec, ""
	if at := strings.Index(keySpec, "@"); at >= 0 {
		field, fieldType = keySpec[:at], keySpec[at+1:]
	}
	field = strings.TrimSpace(field)
	if field == "" {
		fmt.Printf("[pbtool] 忽略无效 dbpool 表规则: %s (缺少分片字段)\n", tableSpec)
		return nil
	}
	return &DBShard{
		Table:     table,
		Shard:     shard,
		ShardKey:  field,
		GoField:   toGoExportedField(field),
		ShardType: strings.TrimSpace(fieldType),
	}
}

// toGoExportedField 将注解中的字段名转为 Go 导出字段名（首字母大写）
// uid -> Uid; Uid -> Uid
func toGoExportedField(field string) string {
	if field == "" {
		return field
	}
	return strings.ToUpper(field[:1]) + field[1:]
}

// hasMember 判断结构体是否包含指定名称的字段
func hasMember(desc TypeDescriptor, field string) bool {
	for _, m := range desc.Members() {
		if m.Name == field {
			return true
		}
	}
	return false
}

// GetAllDBShard 返回声明了 @pbtool dbpool 分库分表规则的结构体
func (p *Parser) GetAllDBShard() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if item.DBShard != nil {
			rets = append(rets, item)
		}
	}
	return
}

func (p *Parser) GetPkgName() string {
	return p.pkgName
}

func (p *Parser) GetAllEnum() []*EnumDescriptor {
	return p.enums
}

func (p *Parser) GetAllStruct() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if strings.HasSuffix(item.Name, "Rsp") ||
			strings.HasSuffix(item.Name, "Req") ||
			strings.HasSuffix(item.Name, "ConfigS") ||
			strings.HasSuffix(item.Name, "Config") ||
			strings.HasSuffix(item.Name, "ConfigAry") {
			continue
		}
		rets = append(rets, item)
	}
	return
}

func (p *Parser) GetAllRsp() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if strings.HasSuffix(item.Name, "Rsp") {
			rets = append(rets, item)
		}
	}
	return
}

// GetAllConfig 返回所有以 Config 结尾（但不以 ConfigAry 或 ConfigS 结尾）的配置结构体
func (p *Parser) GetAllConfig() (rets []*StructDescriptor) {
	for _, item := range p.list {
		if strings.HasSuffix(item.Name, "Config") &&
			!strings.HasSuffix(item.Name, "ConfigAry") &&
			!strings.HasSuffix(item.Name, "ConfigS") {
			rets = append(rets, item)
		}
	}
	return
}

// isRewardType 递归检查类型描述符是否为 *Reward 或 []*Reward
func isRewardType(desc TypeDescriptor) (isReward bool, isSlice bool) {
	inner := desc
	for {
		switch inner.Kind() {
		case KindSlice:
			isSlice = true
			elems := inner.Elements()
			if len(elems) == 0 {
				return false, false
			}
			inner = elems[0]
		case KindPointer:
			elems := inner.Elements()
			if len(elems) == 0 {
				return false, false
			}
			inner = elems[0]
		default:
			// 检查选择器类型 (pkg.Reward)
			if inner.Kind() == KindSelector {
				if sel, ok := inner.(*SelectorTypeDescriptor); ok {
					return sel.Sel == "Reward", isSlice
				}
			}
			// 检查裸标识符类型 (Reward，无包前缀)
			if inner.Kind() == KindBasic {
				if ident, ok := inner.(*IdentTypeDescriptor); ok {
					return ident.IdentName == "Reward", isSlice
				}
			}
			return false, false
		}
	}
}

func (p *Parser) Gen(dst string) error {
	funcMap := template.FuncMap{
		"hasSuffix":     strings.HasSuffix,
		"isExported":    func(name string) bool { return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' },
		"isRewardField": func(m Member) bool { ok, _ := isRewardType(m.Type); return ok },
		"isRewardSlice": func(m Member) bool { _, ok := isRewardType(m.Type); return ok },
		"memberType":    func(m Member) string { return m.Type.Name() },
	}
	tplObj := template.Must(template.New("pb").Funcs(funcMap).Parse(templ))
	buf := bytes.NewBuffer(nil)
	tplObj.Execute(buf, p)
	return futil.Save(filepath.Join(dst, "common.gen.pb.go"), buf.Bytes())
}
