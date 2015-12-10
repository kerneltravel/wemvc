package wemvc

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

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

func titleCase(src string) string {
	if len(src) <= 1 {
		return src
	}
	return strings.ToUpper(string(src[0:1])) + strings.ToLower(string(src[1:]))
}

// Md5String get the md5 code of the string
func Md5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func guidRand() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return Md5String(base64.URLEncoding.EncodeToString(b))
}

// NewGUID make a GUID String
func NewGUID() string {
	if runtime.GOOS == "windows" {
		guid := guidRand()
		return guid
	}
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return guidRand()
	}
	defer f.Close()

	b := make([]byte, 16)
	_, err = f.Read(b)
	if err != nil {
		return guidRand()
	}
	uuid := fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}
