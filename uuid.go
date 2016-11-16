package wemvc

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"regexp"
	"errors"
	"encoding/hex"
)

// UUID define the uuid
type UUID [16]byte

// String print the uuid as long string like '{xxxxxxxx-xxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx}'
func (uuid UUID) String() string {
	bytes := [16]byte(uuid)
	str := fmt.Sprintf("{%x-%x-%x-%x-%x}", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	return strings.ToUpper(str)
}

// ShortString print the uuid as short string like 'xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'
func (uuid UUID) ShortString() string {
	bytes := [16]byte(uuid)
	str := fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
	return strings.ToUpper(str)
}

func (uuid UUID) Equal(newUUid UUID) bool {
	for i := 0; i < 16; i++ {
		if uuid[i] != newUUid[i] {
			return false
		}
	}
	return true
}

func uuidRandBytes() UUID {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return UUID{0}
	}
	bytes := md5Bytes(base64.URLEncoding.EncodeToString(b))
	if len(bytes) != 16 {
		return UUID{0}
	}
	var uuidBytes [16]byte
	copy(uuidBytes[:], bytes)
	return UUID(uuidBytes)
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

	b := []byte{}
	_, err = f.Read(b)
	if err != nil || len(b) != 16 {
		return uuidRandBytes()
	}
	uuid := UUID{}
	copy(uuid[:], b)

	return uuid
}

var uuidRegex *regexp.Regexp = regexp.MustCompile(`^([a-fA-F0-9]{32}|[a-fA-F0-9]{8}(-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}|\{[a-fA-F0-9]{8}(-[a-fA-F0-9]{4}){3}-[a-fA-F0-9]{12}\})$`)

func ParseUUID(s string) (id UUID, err error) {
	if len(s) == 0 {
		err = errors.New("Empty UUID string")
		return
	}

	parts := uuidRegex.FindStringSubmatch(s)
	if parts == nil {
		err = errors.New("Invalid UUID string format")
		return
	}
	var array [16]byte
	slice, _ := hex.DecodeString(strings.Join(parts[1:], ""))
	copy(array[:], slice)
	id = array
	return
}