package wemvc

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"strings"
	"unsafe"
)

func md5Bytes(s string) []byte {
	h := md5.New()
	h.Write(str2Byte(s))
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

func byte2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func str2Byte(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0],x[1],x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func strAdd(arr ...string) string {
	return strings.Join(arr, "")
}