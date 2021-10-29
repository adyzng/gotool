package tpl

import (
	"fmt"
	"os"
	"testing"
	"text/template"
)

func convValue(val interface{}, tmp string) error {
	fmt.Printf("val=%+v tmp=%s", val, tmp)
	return nil
}

func TestMap2StructTpl(t *testing.T) {
	tpl, err := template.New("map2struct").Parse(MapToStructTemplate)
	if err != nil {
		t.Errorf("parse tpl failed. err=%v", err)
		return
	}

	tplInst := tpl.Funcs(template.FuncMap{
		"conv": convValue,
	})

	data := MapToStructTemplateData{
		Package:   "test",
		ParamName: "mp",
		ParamType: "map[string]string",
		ModelPkg:  "model",
		ModelName: "ApiBookItem",
		DirectFields: []*FieldItem{
			{
				JsonName:  "id",
				FieldName: "Id",
				TypeEqual: true,
			},
			{
				JsonName:  "type",
				FieldName: "Type",
				TypeEqual: true,
			},
		},
		AssignFields: []*FieldItem{
			{
				JsonName:   "book_name",
				FieldName:  "BookName",
				AssignExpr: "cast.ToString",
			},
			{
				JsonName:   "book_title",
				FieldName:  "BookTitle",
				AssignExpr: "&",
			},
		},
		OtherFields: []*FieldItem{
			{
				JsonName:  "unknown",
				FieldName: "unknown",
			},
			{
				JsonName:  "unknown2",
				FieldName: "unknown2",
			},
		},
	}

	err = tplInst.Execute(os.Stdout, &data)
	t.Logf("error=%v", err)
}
