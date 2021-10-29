package parse

import (
	"go/ast"
	"reflect"

	"github.com/adyzng/gotool/utils"
)

type TypeKind int

const (
	Unknown TypeKind = iota
	Pointer
	Struct
	Array
)

type StructV2 struct {
	Name    string
	PkgName string
	Package *PackageV2
	AstInfo *ast.StructType `json:"-"`
}

type TypeInfo struct {
	Name     string
	JsonName string
	Package  string
	Type     string
	Kind     TypeKind
}

func (si *StructV2) EnumField(iter func(fd *ast.Field) bool) {
	for _, fd := range si.AstInfo.Fields.List {
		if iter != nil && !iter(fd) {
			break
		}
	}
}

func (si *StructV2) FieldName(fd *ast.Field) string {
	if len(fd.Names) == 1 {
		return fd.Names[0].Name
	} else {
		return ""
	}
}

func (si *StructV2) JsonName(fd *ast.Field) string {
	if fd.Tag == nil {
		return ""
	}
	return utils.GetJsonTagName(reflect.StructTag(fd.Tag.Value).Get("json"))
}

func (si *StructV2) FieldType(fd *ast.Field) *TypeInfo {
	ti := &TypeInfo{
		Kind:     Unknown,
		Name:     si.FieldName(fd),
		JsonName: si.JsonName(fd),
	}

	switch t := fd.Type.(type) {
	case *ast.Ident:
		ti.Type = t.Name

	case *ast.StarExpr:
		ti.Kind = Pointer
		switch st := t.X.(type) {
		case *ast.Ident:
			ti.Type = st.Name

		case *ast.SelectorExpr:
			ti.Type = st.Sel.Name
			ti.Package = st.X.(*ast.Ident).Name
		}

	case *ast.SelectorExpr:
		ti.Type = t.Sel.Name
		ti.Package = t.X.(*ast.Ident).Name

		//case *ast.ArrayType:
		//	ti.Kind = Array
		//	switch elt := t.Elt.(type) {
		//	case *ast.Ident:
		//		ti.Type = elt.Name
		//	}
	}

	if ti.Package == "" && !utils.IsBaseType(ti.Type) {
		ti.Package = si.PkgName
	}
	return ti
}

func (ti *TypeInfo) IsBaseType() bool {
	return utils.IsBaseType(ti.Type)
}

func (ti *TypeInfo) IsObjectType() bool {
	if ti.Package != "" && ti.Type != "" && !utils.IsBaseType(ti.Type) {
		return true
	}
	return false
}
