package wemvc

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type UUID []byte

func (uuid UUID) String() string {
	if len(uuid) != 16 {
		return ""
	}
	bytes := []byte(uuid)
	var s = fmt.Sprintf("{%x-%x-%x-%x-%x}", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	return strings.ToUpper(s)
}

func (uuid UUID) ShortString() string {
	if len(uuid) != 16 {
		return ""
	}
	bytes := []byte(uuid)
	var s = fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	return strings.ToUpper(s)
}

func uuidRandBytes() UUID {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil
	}
	var uuidBytes = md5Bytes(base64.URLEncoding.EncodeToString(b))
	return uuidBytes
}

// NewUUID make a UUID String
func NewUUID() UUID {
	if runtime.GOOS == "windows" {
		uuid := uuidRandBytes()
		return uuid
	}
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return uuidRandBytes()
	}
	defer f.Close()

	b := make([]byte, 16)
	_, err = f.Read(b)
	if err != nil {
		return uuidRandBytes()
	}
	return b
}
