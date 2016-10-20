package wemvc

import (
	"os"
	"runtime"
	"strings"
)

func fixPath(src string) string {
	var res string
	if runtime.GOOS == `windows` {
		res = strings.Replace(src, "/", "\\", -1)
	} else {
		res = strings.Replace(src, "\\", "/", -1)
	}
	return res
}

// IsDir check if the path is directory
func IsDir(path string) bool {
	state, err := os.Stat(path)
	if err != nil {
		return false
	}
	return state.IsDir()
}

// IsFile check if the path is file
func IsFile(path string) bool {
	state, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !state.IsDir()
}

// WorkingDir get the current working directory
func WorkingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
