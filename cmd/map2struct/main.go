package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"

	"github.com/adyzng/gotool/parse"
	"github.com/adyzng/gotool/tpl"
	"github.com/adyzng/gotool/utils"
)

var (
	input  = flag.String("input", "", "input file path")
	output = flag.String("output", "", "output file path; default ./<input>_gen.go")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of map2struct:\n")
	fmt.Fprintf(os.Stderr, "    map2struct [flags] -input=xx -output=xx\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "    https://github.com/adyzng/gotool/blob/master/cmd/map2struct/README.md\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("map2struct: ")
	flag.Usage = Usage
	flag.Parse()

	inputFile := *input
	outputFile := *output

	if inputFile == "" {
		inputFile = os.Getenv("GOFILE")
		inputFile = strings.Trim(inputFile, `"`)
	}
	if inputFile == "" {
		log.Fatal("invalid param.")
		return
	}

	if !filepath.IsAbs(inputFile) {
		cwd, _ := os.Getwd()
		inputFile = filepath.Join(cwd, inputFile)
	}

	if outputFile == "" {
		outputFile = filepath.Dir(inputFile)
	} else {
		outputFile = strings.Trim(outputFile, `"`)
	}

	cwd, _ := os.Getwd()
	outputFile = filepath.Join(outputFile, newOutputName(inputFile))
	log.Printf("cwd: %s", cwd)
	log.Printf("input: %s", inputFile)
	log.Printf("output: %s", outputFile)

	pkgParser := parse.NewPkgParser()
	genPkg, err := pkgParser.ParseFile(inputFile)
	if err != nil {
		log.Fatalf("parse function failed. path=%s err=%v", inputFile, err)
		return
	}

	fnList, err := genPkg.FindFuncList("MapTo")
	if err != nil {
		log.Fatalf("parse function failed. path=%s err=%v", inputFile, err)
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	if err = processPrefix(genPkg, buffer); err != nil {
		log.Fatalf("process package failed. err=%v", err)
		return
	}

	for _, fun := range fnList {
		log.Printf(
			"➡️ func=%s ing, in=map[%s]%s, out=%s.%s",
			fun.Name, fun.InputParam.KeyType, fun.InputParam.ValueType,
			fun.OutputType.Package, fun.OutputType.TypeName,
		)
		//if fun.OutputParam == nil {
		//	log.Printf("function first return value is not struct. func=%v", fun)
		//	continue
		//}
		if err = map2Struct(genPkg, fun, buffer); err != nil {
			log.Printf("process function failed. func=%v err=%v", fun, err)
			return
		}
		log.Printf("✅ func=%s done", fun.Name)
	}

	// log.Printf("%s", buffer.String())
	if err = saveOutput(buffer.Bytes(), outputFile); err != nil {
		log.Printf("%s", buffer.Bytes())
		log.Fatalf("save output failed. err=%v", err)
		return
	}

	log.Printf("succeed")
	return
}

func map2Struct(pkg *parse.PackageV2, fun *parse.FunctionV2, writer io.Writer) (err error) {
	tplData := tpl.MapToStructTemplateData{
		FuncName:     "gen" + fun.Name,
		ParamName:    "src",
		ParamType:    fmt.Sprintf("map[%s]%s", fun.InputParam.KeyType, fun.InputParam.ValueType),
		ModelName:    fun.OutputType.TypeName,
		ModelPkg:     fun.OutputType.Package,
		EnumFields:   []*tpl.FieldItem{},
		DirectFields: []*tpl.FieldItem{}, // 类型相同
		AssignFields: []*tpl.FieldItem{}, // optional 的字段，
		OtherFields:  []*tpl.FieldItem{}, // 类型不同的，非optional字段
	}

	input := fun.InputParam
	model := fun.OutputParam
	model.EnumField(func(fd *ast.Field) bool {
		ft := model.FieldType(fd)
		fdItem := &tpl.FieldItem{
			JsonName:  ft.JsonName,
			FieldName: ft.Name,
			FieldType: ft.Type,
			IsPointer: ft.Kind == parse.Pointer,
			TypeEqual: ft.Type == input.ValueType,
		}

		switch {
		case ft.IsBaseType(): // 基本类型
			if ft.Kind == parse.Pointer {
				fdItem.GenType = "assign"
			} else {
				fdItem.GenType = "direct"
			}
			if !fdItem.TypeEqual || input.IsValueInterface {
				fdItem.AssignExpr = fmt.Sprintf("cast.To%s", utils.ToCap(ft.Type))
			}

		case ft.IsObjectType(): // 可能是枚举
			depPkg, err := pkg.GetImportPkg(ft.Package)
			if err != nil || depPkg == nil {
				log.Printf("process field failed=%s.%s, err=%v", model.Name, ft.JsonName, err)
				return false
			}
			if ti := depPkg.GetTypeIdent(ft.Type); utils.IsBaseType(ti.Type) {
				fdItem.GenType = "enum"
				fdItem.TypeConv = fmt.Sprintf("%s.%s", ft.Package, ft.Type)
				fdItem.AssignExpr = fmt.Sprintf("cast.To%s", utils.ToCap(ti.Type))
			}
		}

		switch fdItem.GenType {
		case "enum":
			tplData.EnumFields = append(tplData.EnumFields, fdItem)
			log.Printf("enum field. field=%s.%s type=%+v", model.Name, ft.Name, ft)
		case "direct":
			tplData.DirectFields = append(tplData.DirectFields, fdItem)
			log.Printf("direct field. field=%s.%s type=%+v", model.Name, ft.Name, ft)
		case "assign":
			tplData.AssignFields = append(tplData.AssignFields, fdItem)
			log.Printf("assign field. field=%s.%s type=%+v", model.Name, ft.Name, ft)
		default:
			tplData.OtherFields = append(tplData.OtherFields, fdItem)
			log.Printf("⚠️ unknown field. field=%s.%s type=%+v", model.Name, ft.Name, ft)
		}

		return true
	})

	tplInst := template.New("mapToStruct")
	if tplInst, err = tplInst.Parse(tpl.MapToStructTemplate); err != nil {
		log.Printf("template parsed failed. err=%v", err)
		return err
	}
	if err := tplInst.Execute(writer, &tplData); err != nil {
		log.Printf("template exceute failed. err=%v", err)
		return err
	}
	return nil
}

func processPrefix(pkg *parse.PackageV2, writer io.Writer) (err error) {
	tplData := tpl.MapToStructTemplateData{
		Package: pkg.Name,
	}

	tplInst := template.New("mapToStruct")
	if tplInst, err = tplInst.Parse(tpl.MapToStructPrefix); err != nil {
		log.Printf("template parsed failed. err=%v", err)
		return err
	}

	if err := tplInst.Execute(writer, &tplData); err != nil {
		log.Printf("template exceute failed. err=%v", err)
		return err
	}

	return nil
}

func saveOutput(data []byte, file string) error {
	out, err := imports.Process(file, data, nil)
	if err != nil {
		log.Printf("import error. err=%v", err)
		return err
	}
	return ioutil.WriteFile(file, out, 0600)
}

func newOutputName(file string) string {
	fname := filepath.Base(file)
	fname = strings.TrimSuffix(fname, ".go")
	return utils.ToSnakeCase(fname) + "_gen.go"
}
