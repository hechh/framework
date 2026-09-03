package internal

const templ = `
/*
* 本代码由pbtool工具生成，请勿手动修改
*/

package {{.GetPkgName}}

{{range $st := .GetAllDBShard -}}
func (d *{{$st.Name}}) TableName() string {
	return fmt.Sprintf("{{$st.DBShard.Table}}_%d", d.{{$st.DBShard.GoField}}%{{$st.DBShard.Shard}})
}

{{end}}
{{range $st := .GetAllRsp -}}
{{range $field := $st.Members -}}
{{if hasSuffix $field.Type.Name ".RspHead" -}}
func (d *{{$st.Name}}) SetRspHead(code int32, msg string) {
	d.{{$field.Name}} = &{{index $field.Type.Elements 0}}{Code:code, Msg:msg}
}

func (d *{{$st.Name}}) GetRspHead() (int32, string) {
	if d.{{$field.Name}} != nil {
		return d.{{$field.Name}}.Code, d.{{$field.Name}}.Msg
	}
	return 0, ""	
}
{{end}}
{{end}}
{{end}}

{{range $st := .GetAllStruct -}}
func(d *{{$st.Name}}) ToDB() ([]byte, error) {
	if d == nil {
		return nil, nil
	}
	return d.MarshalVT()
}

func(d *{{$st.Name}}) FromDB(val []byte) error {
	if len(val) <= 0 {
		return nil
	}
	return d.UnmarshalVT(val)
}

{{end}}

{{range $cfg := .GetAllConfig -}}
// {{$cfg.Name}}S 是 {{$cfg.Name}} 的只读包装，通过 Get 方法安全访问内部数据。
// Reward 及 []*Reward 类型字段统一返回 CloneVT 深拷贝，避免业务层修改缓存配置
// （历史教训：GetXxxReadOnly 返回内部指针，被 reward.Merge/Reward 原地修改污染配置表）。
type {{$cfg.Name}}S struct {
	inner *{{$cfg.Name}}
}

// ToS 将 {{$cfg.Name}} 转换为只读包装
func (d *{{$cfg.Name}}) ToS() *{{$cfg.Name}}S {
	return &{{$cfg.Name}}S{inner: d}
}

{{range $field := $cfg.Members -}}
{{if isExported $field.Name -}}
{{if isRewardField $field -}}
{{if isRewardSlice $field -}}
func (s *{{$cfg.Name}}S) Get{{$field.Name}}() {{memberType $field}} {
	if s.inner.{{$field.Name}} == nil {
		return nil
	}
	src := &RewardList{List: s.inner.{{$field.Name}}}
	dst := src.CloneVT()
	return dst.List
}

{{else -}}
func (s *{{$cfg.Name}}S) Get{{$field.Name}}() {{memberType $field}} {
	return s.inner.{{$field.Name}}.CloneVT()
}

{{end}}
{{else -}}
func (s *{{$cfg.Name}}S) Get{{$field.Name}}() {{memberType $field}} {
	return s.inner.{{$field.Name}}
}

{{end}}
{{end}}
{{end}}
{{end}}
`
