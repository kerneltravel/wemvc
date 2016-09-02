package wemvc

import (
	"crypto/md5"
	"encoding/hex"
)

func md5Bytes(s string) []byte {
	h := md5.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

// Md5String get the md5 code of the string
func Md5String(s string) string {
	return hex.EncodeToString(md5Bytes(s))
}