package parse

import (
	"encoding/json"
	"go/ast"
	"os"
	"path/filepath"
	"testing"
)

func toJsonStr(obj interface{}) string {
	ds, _ := json.MarshalIndent(obj, "", "  ")
	return string(ds)
}

func TestParsePackage(t *testing.T) {
	dir := "/Users/huan/go/src/github.com/adyzng/map2struct/cmd/map2struct"
	pkgPath := "github.com/adyzng/xxx/kitex_gen/comic_base"

	_ = os.Chdir(dir)

	pp := NewPkgParser()
	pkg, err := pp.ParsePackage(pkgPath)
	t.Logf("err=%v pkg=%s", err, toJsonStr(pkg))

	subPkg, err := pkg.GetImportPkg("bthrift")
	t.Logf("err=%v pkg=%+v", err, toJsonStr(subPkg))

	st, err := pkg.FindStruct("ApiBookInfo")
	t.Logf("err=%v st=%+v", err, toJsonStr(st))

	st.EnumField(func(fd *ast.Field) bool {
		typ := st.FieldType(fd)
		t.Logf("%+v", typ)
		return true
	})
}

func TestParseFile(t *testing.T) {
	filePath := "/Users/huan/go/src/github.com/adyzng/xxx/utils/convert/convert.go"
	fileDir := filepath.Dir(filePath)

	_ = os.Chdir(fileDir)

	curMod, _ := GetModDetail("")
	t.Logf("mod=%s", toJsonStr(curMod))

	pp := NewPkgParser()
	pkg, err := pp.ParseFile(filePath)
	t.Logf("err=%v pkg=%s", err, toJsonStr(pkg))

	depPkg, err := pkg.GetImportPkg("comic_base")
	t.Logf("err=%v pkg=%+v", err, toJsonStr(depPkg))

	st, err := depPkg.FindStruct("ApiItemInfo")
	t.Logf("err=%v st=%+v", err, toJsonStr(st))
}
