package wemvc

import (
	"encoding/xml"
	"io/ioutil"
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

func file2Xml(fpath string, v interface{}) error {
	bytes, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(bytes, v)
	if err != nil {
		return err
	}
	return nil
}

func isDir(fpath string) bool {
	state, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return state.IsDir()
}

func isFile(fpath string) bool {
	state, err := os.Stat(fpath)
	if err != nil {
		return false
	}
	return !state.IsDir()
}
