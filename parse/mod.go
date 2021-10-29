package parse

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adyzng/toolkit/utils"
)

var (
	cacheMod   = map[string]*Module{}
	stdLibInfo = map[string]string{}
)

func init() {
	src := utils.GetStdSrcPath()
	list, _ := os.ReadDir(src)
	for _, entry := range list {
		if !entry.IsDir() {
			continue
		}
		rootPath := filepath.Join(src, entry.Name())
		stdLibInfo[entry.Name()] = rootPath
	}
}

func IsStdLib(pkgPath string) (string, bool) {
	ss := strings.SplitN(pkgPath, "/", 2)
	if rootPath, ok := stdLibInfo[ss[0]]; ok {
		if len(ss) > 1 {
			rootPath = filepath.Join(rootPath, ss[1])
		}
		return rootPath, true
	}
	return "", false
}

type Module struct {
	Path       string     // module path
	Version    string     // module version
	Versions   []string   // available module versions (with -versions)
	Replace    *Module    // replaced by this module
	Time       *time.Time // time version was created
	Update     *Module    // available update, if any (with -u)
	Main       bool       // is this the main module?
	Indirect   bool       // is this module only an indirect dependency of main module?
	Dir        string     // directory holding files for this module, if any
	GoMod      string     // path to go.mod file for this module, if any
	GoVersion  string     // go version used in module
	Deprecated string     // deprecation message, if any (with -u)
}

func GetCurMod() (*Module, error) {
	return GetModDetail("")
}

func GetModDetail(pkgName string) (*Module, error) {
	runCmd := func(name string) ([]byte, error) {
		cmd := exec.Command("bash", "-c", fmt.Sprintf("go list -m -json %s", name))
		data, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	var err error
	var data []byte
	if cm := cacheMod[pkgName]; cm != nil {
		return cm, nil
	}

	ss := strings.Split(pkgName, "/")
	for n := len(ss); n > 0; n-- {
		name := strings.Join(ss[:n], "/")
		if data, err = runCmd(name); err == nil {
			break
		}
	}
	if len(data) == 0 || err != nil {
		err = fmt.Errorf("unknown mod: %s, (forget to go get ?)", pkgName)
		return nil, err
	}

	m := &Module{}
	if err = json.Unmarshal(data, m); err != nil {
		return nil, err
	}

	cacheMod[pkgName] = m
	return m, nil
}

func GetPkgAbsPath(pkgPath string) string {
	if pkgPath == "C" {
		return ""
	}

	if filepath.IsAbs(pkgPath) {
		return filepath.Clean(pkgPath)
	}

	// 内置库
	if absPath, ok := IsStdLib(pkgPath); ok {
		return absPath
	}

	// 当前 mod 库
	if curMod, _ := GetModDetail(""); curMod != nil {
		if pos := strings.Index(pkgPath, curMod.Path); pos >= 0 {
			tmp := strings.TrimPrefix(pkgPath, curMod.Path)
			return filepath.Join(curMod.Dir, tmp)
		}
	}

	// 第三方库
	mod, err := GetModDetail(pkgPath)
	if err != nil || mod == nil {
		log.Printf("parse mod failed. mod=%s err=%v\n", pkgPath, err)
		panic(err)
		return ""
	}

	// pkg 在 mod 里的子路径
	dir := mod.Dir
	subPath := strings.TrimPrefix(pkgPath, mod.Path)
	if mod.Replace != nil {
		dir = mod.Replace.Dir
	}

	// 绝对路径
	return filepath.Join(dir, subPath)
}
