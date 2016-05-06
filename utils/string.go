package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
	_ "errors"
	_ "fmt"
	"math"
	mrand "math/rand"
	_ "strconv"
	_ "strings"
)


func TitleCase(src string) string {
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


var (
	defaultRand = mrand.New(mrand.NewSource(time.Now().UnixNano()))
)

// Creates a random string based on a variety of options, using
// supplied source of randomness.
//
// If start and end are both 0, start and end are set
// to ' ' and 'z', the ASCII printable
// characters, will be used, unless letters and numbers are both
// false, in which case, start and end are set to 0 and math.MaxInt32.
//
// If set is not nil, characters between start and end are chosen.
//
// This method accepts a user-supplied rand.Rand
// instance to use as a source of randomness. By seeding a single
// rand.Rand instance with a fixed seed and using it for each call,
// the same random sequence of strings can be generated repeatedly
// and predictably.
func RandomSpec0(count uint, start, end int, letters, numbers bool, chars []rune, rand *mrand.Rand) string {
	if count == 0 {
		return ""
	}
	if start == 0 && end == 0 {
		end = 'z' + 1
		start = ' '
		if !letters && !numbers {
			start = 0
			end = math.MaxInt32
		}
	}
	buffer := make([]rune, count)
	gap := end - start
	for count != 0 {
		count--
		var ch rune
		if len(chars) == 0 {
			ch = rune(rand.Intn(gap) + start)
		} else {
			ch = chars[rand.Intn(gap)+start]
		}
		if letters && ((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) ||
			numbers && (ch >= '0' && ch <= '9') ||
			(!letters && !numbers) {
			if ch >= rune(56320) && ch <= rune(57343) {
				if count == 0 {
					count++
				} else {
					buffer[count] = ch
					count--
					buffer[count] = rune(55296 + rand.Intn(128))
				}
			} else if ch >= rune(55296) && ch <= rune(56191) {
				if count == 0 {
					count++
				} else {
					// high surrogate, insert low surrogate before putting it in
					buffer[count] = rune(56320 + rand.Intn(128))
					count--
					buffer[count] = ch
				}
			} else if ch >= rune(56192) && ch <= rune(56319) {
				// private high surrogate, no effing clue, so skip it
				count++
			} else {
				buffer[count] = ch
			}
		} else {
			count++
		}
	}
	return string(buffer)
}

// Creates a random string whose length is the number of characters specified.
//
// Characters will be chosen from the set of alpha-numeric
// characters as indicated by the arguments.
//
// Param count - the length of random string to create
// Param start - the position in set of chars to start at
// Param end   - the position in set of chars to end before
// Param letters - if true, generated string will include
//                 alphabetic characters
// Param numbers - if true, generated string will include
//                 numeric characters
func RandomSpec1(count uint, start, end int, letters, numbers bool) string {
	return RandomSpec0(count, start, end, letters, numbers, nil, defaultRand)
}

// Creates a random string whose length is the number of characters specified.
//
// Characters will be chosen from the set of alpha-numeric
// characters as indicated by the arguments.
//
// Param count - the length of random string to create
// Param letters - if true, generated string will include
//                 alphabetic characters
// Param numbers - if true, generated string will include
//                 numeric characters
func RandomAlphaOrNumeric(count uint, letters, numbers bool) string {
	return RandomSpec1(count, 0, 0, letters, numbers)
}

func RandomString(count uint) string {
	return RandomAlphaOrNumeric(count, false, false)
}

func RandomStringSpec0(count uint, set []rune) string {
	return RandomSpec0(count, 0, len(set)-1, false, false, set, defaultRand)
}

func RandomStringSpec1(count uint, set string) string {
	return RandomStringSpec0(count, []rune(set))
}

// Creates a random string whose length is the number of characters
// specified.
//
// Characters will be chosen from the set of characters whose
// ASCII value is between 32 and 126 (inclusive).
func RandomAscii(count uint) string {
	return RandomSpec1(count, 32, 127, false, false)
}

// Creates a random string whose length is the number of characters specified.
// Characters will be chosen from the set of alphabetic characters.
func RandomAlphabetic(count uint) string {
	return RandomAlphaOrNumeric(count, true, false)
}

// Creates a random string whose length is the number of characters specified.
// Characters will be chosen from the set of alpha-numeric characters.
func RandomAlphanumeric(count uint) string {
	return RandomAlphaOrNumeric(count, true, true)
}

// Creates a random string whose length is the number of characters specified.
// Characters will be chosen from the set of numeric characters.
func RandomNumeric(count uint) string {
	return RandomAlphaOrNumeric(count, false, true)
}