package utils

import (
	"path/filepath"
	"runtime"
)

func GetStdSrcPath() string {
	return filepath.Join(runtime.GOROOT(), "src")
}
