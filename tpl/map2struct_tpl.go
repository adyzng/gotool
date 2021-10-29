package tpl

type FieldItem struct {
	GenType    string
	TypeEqual  bool
	IsPointer  bool
	FieldType  string
	FieldName  string
	JsonName   string
	TypeConv   string // 类型转换
	AssignExpr string // 赋值表达式
}

type MapToStructTemplateData struct {
	FuncName  string
	Package   string
	ParamName string
	ParamType string
	ModelPkg  string
	ModelName string

	EnumFields   []*FieldItem // 枚举类型
	DirectFields []*FieldItem // 类型相同
	AssignFields []*FieldItem // 带赋值表达式的，比如：指针类型
	OtherFields  []*FieldItem // 其他不能处理的类型
}

const MapToStructPrefix = `
// Auto generated code, DO NOT EDIT.

package {{.Package}}

import (
	"github.com/spf13/cast"
)
`

const MapToStructTemplate = `

func {{.FuncName}}({{.ParamName}} {{.ParamType}}) (obj *{{.ModelPkg}}.{{.ModelName}}, err error) {
	obj = &{{.ModelPkg}}.{{.ModelName}}{}
	{{$mapParam := .ParamName -}}

	{{ $len1 := len .DirectFields }}
	{{ if gt $len1 0}}
	{{ print "// 直接赋值的字段" }}
	{{- range .DirectFields }}
		{{- if .AssignExpr }}
			{{ printf "obj.%s = %s(%s[\"%s\"])" .FieldName .AssignExpr $mapParam .JsonName }}
		{{- else }}
			{{ printf "obj.%s = %s[\"%s\"]" .FieldName $mapParam .JsonName }}
		{{- end }}
	{{- end }}
	{{- end -}}

	{{ $len2 := len .EnumFields }}
	{{ if gt $len2 0}}
		{{ print "// 枚举类型" }}
		{{- range .EnumFields }}
			{{ printf "if tmp, ok := %s[\"%s\"]; ok {" $mapParam .JsonName }}
				{{- if .AssignExpr }}
					{{ printf "	val := (%s)(%s(tmp))" .TypeConv .AssignExpr }}
					{{- if .IsPointer }}
						{{ printf "	obj.%s = &val" .FieldName }}
					{{- else }}
						{{ printf "	obj.%s = val" .FieldName }}
					{{- end }}
				{{- else }}
					{{- if .IsPointer }}
						{{ printf "	obj.%s = &tmp" .FieldName }}
					{{- else }}
						{{ printf "	obj.%s = tmp" .FieldName }}
					{{- end }}
				{{- end }}
			{{ print "}" }}
		{{- end }}
	{{- end -}}

	{{ $len2 := len .AssignFields }}
	{{ if gt $len2 0}}
	{{ print "// 带赋值表达式的（指针类型）" }}
	{{- range .AssignFields }}
		{{ printf "if tmp, ok := %s[\"%s\"]; ok {" $mapParam .JsonName }}
			{{- if .AssignExpr }}
				{{ printf "	val := %s(tmp)" .AssignExpr }}
				{{- if .IsPointer }}
					{{ printf "	obj.%s = &val" .FieldName }}
				{{- else }}
					{{ printf "	obj.%s = val" .FieldName }}
				{{- end }}
			{{- else }}
				{{- if .IsPointer }}
					{{ printf "	obj.%s = &tmp" .FieldName }}
				{{- else }}
					{{ printf "	obj.%s = tmp" .FieldName }}
				{{- end }}
			{{- end }}
		{{ print "}" }}
	{{- end }}
	{{- end -}}

	{{ $len3 := len .OtherFields }}
	{{ if gt $len3 0}}
		{{ print "// 需要手动处理的字段" }}
		{{- range .OtherFields }}
			{{ printf "// obj.%s = ?" .FieldName }}
		{{- end }}
	{{- end }}

	return obj, err
}
`
