package parse

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

type PkgParser interface {
	ParseFile(filePath string) (*PackageV2, error)
	ParsePackage(pkgPath string) (*PackageV2, error)
}

type PackageV2 struct {
	Name       string            // 包名
	AliasName  string            // import别名
	PkgPath    string            // import路径
	SourcePath string            // 文件绝对路径
	PackageAst *ast.Package      `json:"-"`
	Imports    map[string]string // key:import名, value:import path

	parser PkgParser `json:"-"`
}

type DelayPkgParser struct {
	depth      int
	allPkgInfo map[string]*PackageV2 // key: import 路径
}

func NewPkgParser() PkgParser {
	return &DelayPkgParser{
		depth:      2,
		allPkgInfo: map[string]*PackageV2{},
	}
}

func (pp *DelayPkgParser) filter(fi fs.FileInfo) bool {
	if strings.HasSuffix(fi.Name(), "_test.go") {
		return false
	}
	return true
}

func (pp *DelayPkgParser) astParseDir(dir string, mode parser.Mode) (string, *ast.Package, error) {
	fSet := token.NewFileSet()
	pkgMap, err := parser.ParseDir(fSet, dir, pp.filter, mode)
	if err != nil {
		return "", nil, err
	}
	for name, pkg := range pkgMap {
		return name, pkg, nil
	}
	return "", nil, errors.New("should not be here")
}

func (pp *DelayPkgParser) parseDir(srcDir string, pkgPath string) (*PackageV2, error) {
	pkgName, pkgAst, err := pp.astParseDir(srcDir, 0)
	if err != nil || pkgAst == nil {
		log.Fatalf("parse dir failed. dir=%s err=%v", srcDir, err)
		return nil, err
	}

	pv := &PackageV2{
		Name:       pkgName,
		PackageAst: pkgAst,
		PkgPath:    pkgPath,
		SourcePath: srcDir,
		parser:     pp,
		Imports:    map[string]string{},
	}
	//if _, ok := IsStdLib(pkgPath); ok {
	//	pv.StdLib = true
	//}

	ast.Inspect(pkgAst, func(node ast.Node) bool {
		switch idt := node.(type) {
		case *ast.ImportSpec:
			pkgPath := strings.Trim(idt.Path.Value, "\"")
			pkgName := pp.ParsePkgName(pkgPath)
			pv.Imports[pkgName] = pkgPath

			// 别名
			if idt.Name != nil && idt.Name.Name != "" {
				aliasName := idt.Name.Name
				pv.Imports[aliasName] = pkgPath
			}
			//log.Printf("pkg=%s depend=%s", pv.Name, pkgPath)
		}
		return true
	})

	pp.allPkgInfo[pv.PkgPath] = pv
	return pv, nil
}

// pkgPath 包引用路径，非绝对路径
func (pp *DelayPkgParser) ParsePkgName(pkgPath string) string {
	dir := GetPkgAbsPath(pkgPath)
	name, _, err := pp.astParseDir(dir, parser.PackageClauseOnly)
	if err != nil {
		log.Fatalf("parse package name failed. dir=%s err=%v", pkgPath, err)
	}

	if name == "" {
		ss := strings.Split(pkgPath, "/")
		name = ss[len(ss)-1]
	}
	return name
}

// pkgPath 包引用路径，非绝对路径
func (pp *DelayPkgParser) ParsePackage(pkgPath string) (*PackageV2, error) {
	if pkgPath == "" {
		return nil, nil
	}
	if pkg, ok := pp.allPkgInfo[pkgPath]; ok {
		return pkg, nil
	}

	pkgDir := GetPkgAbsPath(pkgPath)
	return pp.parseDir(pkgDir, pkgPath)
}

// filePath 文件绝对路径
func (pp *DelayPkgParser) ParseFile(filePath string) (*PackageV2, error) {
	//fileName := filepath.Base(filePath)
	return pp.parseDir(filepath.Dir(filePath), filepath.Base(filePath))
}

func (p *PackageV2) GetImportPkg(pkgName string) (*PackageV2, error) {
	pkgPath := p.Imports[pkgName]
	if pkgPath == "" {
		return nil, fmt.Errorf("unknown import package. pkgPath=%s import=%s", pkgPath, pkgName)
	}
	return p.parser.ParsePackage(pkgPath)
}

func (p *PackageV2) FindFuncList(prefix string) (map[string]*FunctionV2, error) {
	if p.PackageAst == nil {
		return nil, fmt.Errorf("package not parsed. pkg=%s", p.Name)
	}

	fnList := map[string]*FunctionV2{}
	ast.Inspect(p.PackageAst, func(node ast.Node) bool {
		switch idt := node.(type) {
		case *ast.FuncDecl:
			fi, err := p.parseFuncDecl(idt, prefix)
			if fi == nil || err != nil {
				return true
			} else {
				fnList[fi.Name] = fi
			}
		}
		return true
	})
	return fnList, nil
}

func (p *PackageV2) FindStruct(stName string) (*StructV2, error) {
	if p.PackageAst == nil {
		return nil, fmt.Errorf("package not parsed. pkg=%s", p.Name)
	}

	var si *StructV2
	ast.Inspect(p.PackageAst, func(node ast.Node) bool {
		switch spec := node.(type) {
		case *ast.TypeSpec:
			st, ok := spec.Type.(*ast.StructType)
			if !ok || stName != spec.Name.Name {
				return true
			}
			si = &StructV2{
				Name:    stName,
				PkgName: p.Name,
				Package: p,
				AstInfo: st,
			}
			// log.Printf("find function struct: %s.%s", si.PkgName, si.Name)
			return false
		}
		return true
	})
	return si, nil
}

func (p *PackageV2) parseStructField(idt *ast.Field) *ObjectType {
	ot := &ObjectType{}
	if len(idt.Names) >= 1 {
		ot.Name = idt.Names[0].Name
	}
	switch t := idt.Type.(type) {
	case *ast.Ident:
		ot.Package = p.Name
		ot.TypeName = t.Name
	case *ast.StarExpr:
		ot.Pointer = true
		switch st := t.X.(type) {
		case *ast.Ident:
			ot.Package = p.Name
			ot.TypeName = st.Name
		case *ast.SelectorExpr:
			x := st.X.(*ast.Ident)
			ot.Package = x.Name
			ot.TypeName = st.Sel.Name
		}
	case *ast.SelectorExpr:
		x := t.X.(*ast.Ident)
		ot.Package = x.Name
		ot.TypeName = t.Sel.Name
	case *ast.InterfaceType:
		ot.TypeName = "interface"
	default:
		log.Printf("unhandled type: %+v(%T)", t, t)
	}
	return ot
}

func (p *PackageV2) GetTypeIdent(idtName string) *TypeInfo {
	typ := &TypeInfo{
		Package: p.Name,
	}
	if idtName == "" || p.PackageAst == nil {
		return typ
	}
	ast.Inspect(p.PackageAst, func(node ast.Node) bool {
		switch nt := node.(type) {
		case *ast.TypeSpec:
			if idtName != nt.Name.Name {
				return true
			}
			switch ts := nt.Type.(type) {
			case *ast.StructType: // struct
				typ.Kind = Struct
			case *ast.Ident: // enum
				typ.Type = ts.Name
			}
			return false
		}
		return true
	})
	return typ
}

func (p *PackageV2) parseFuncDecl(idt *ast.FuncDecl, prefix string) (*FunctionV2, error) {
	var err error
	if idt == nil || idt.Type == nil {
		return nil, err
	}
	if prefix != "" && !strings.HasPrefix(idt.Name.Name, prefix) {
		return nil, err
	}

	params := idt.Type.Params
	returns := idt.Type.Results
	if len(params.List) < 1 || len(returns.List) < 1 {
		log.Printf("not supported function. decl=%#v", idt)
		return nil, errors.New("not supported function signature")
	}

	fi := &FunctionV2{
		Name:    idt.Name.Name,
		FuncAst: idt.Type,
	}

	// 找到第一个map类型
	for idx, arg := range params.List {
		if pArg := parseMapType(arg); pArg != nil {
			fi.InputParam = pArg
			break
		} else {
			pTyp := p.parseStructField(arg)
			log.Printf("args[%d]=%+v", idx, pTyp)
		}
	}

	fi.OutputType = p.parseStructField(returns.List[0])
	depPkg, err := p.GetImportPkg(fi.OutputType.Package)
	if depPkg != nil {
		fi.OutputParam, err = depPkg.FindStruct(fi.OutputType.TypeName)
		if err != nil {
			log.Printf("parse struct failed. struct=%+v", fi.OutputType)
			return nil, err
		}
	}

	log.Printf("found function. fun=%s input=%+v output=%+v", fi.Name, fi.InputParam, fi.OutputType)
	return fi, err
}
