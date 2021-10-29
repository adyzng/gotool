package parse

import (
	"go/ast"
	"log"
)

type FunctionV2 struct {
	Name        string
	FuncAst     *ast.FuncType `json:"-"`
	InputParam  *MapType
	OutputType  *ObjectType
	OutputParam *StructV2
}

type MapType struct {
	KeyType          string
	ValueType        string
	IsValueInterface bool
}

type ObjectType struct {
	Name     string
	Pointer  bool
	Package  string
	TypeName string
}

func parseMapType(idt *ast.Field) *MapType {
	switch t := idt.Type.(type) {
	case *ast.MapType:
		key := t.Key.(*ast.Ident)
		mt := &MapType{
			KeyType: key.Name,
		}

		switch vt := t.Value.(type) {
		case *ast.Ident:
			mt.ValueType = vt.Name
		case *ast.InterfaceType:
			mt.ValueType = "interface{}"
			mt.IsValueInterface = true
		}
		return mt
	}
	return nil
}

func parseObjectType(idt *ast.Field, pkg *PackageV2) *ObjectType {
	ot := &ObjectType{}
	if len(idt.Names) >= 1 {
		ot.Name = idt.Names[0].Name
	}
	switch t := idt.Type.(type) {
	case *ast.Ident:
		ot.Package = pkg.Name
		ot.TypeName = t.Name
		return ot
	case *ast.StarExpr:
		ot.Pointer = true
		switch st := t.X.(type) {
		case *ast.Ident:
			ot.Package = pkg.Name
			ot.TypeName = st.Name
		case *ast.SelectorExpr:
			x := st.X.(*ast.Ident)
			ot.Package = x.Name
			ot.TypeName = st.Sel.Name
		}
		return ot
	case *ast.SelectorExpr:
		x := t.X.(*ast.Ident)
		ot.Package = x.Name
		ot.TypeName = t.Sel.Name
		return ot
	case *ast.InterfaceType:
		ot.TypeName = "interface"
		return ot
	default:
		log.Printf("unhandled type: %+v(%T)", t, t)
	}
	return nil
}
