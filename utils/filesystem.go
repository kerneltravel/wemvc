package utils

import (
	"path/filepath"
	"os"
	"strings"
	"log"
	"runtime"
)

func FixPath(src string) string {
	var res string
	if runtime.GOOS == `windows` {
		res = strings.Replace(src, "/", "\\", -1)
	} else {
		res = strings.Replace(src, "\\", "/", -1)
	}
	return res
}

// IsDir check if the path is directory
func IsDir(fpath string) bool {
	state, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return state.IsDir()
}

// IsFile check if the path is file
func IsFile(fpath string) bool {
	state, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return !state.IsDir()
}