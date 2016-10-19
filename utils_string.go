package wemvc

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"strings"
)

func md5Bytes(s string) []byte {
	h := md5.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func getContrllerName(ctrlType reflect.Type) string {
	cName := strings.ToLower(ctrlType.String())
	cName = strings.Split(cName, ".")[1]
	cName = strings.Replace(cName, "controller", "", -1)
	return cName
}

// Md5String get the md5 code of the string
func Md5String(s string) string {
	return hex.EncodeToString(md5Bytes(s))
}
