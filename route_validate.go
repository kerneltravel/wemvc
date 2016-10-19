package wemvc

import (
	"regexp"
	"strings"
)

// RouteFunc define the route check function
type RouteValidateFunc func(urlPath string, opt RouteOption) string

func validateAny(urlPath string, opt RouteOption) string {
	var length = uint8(len(urlPath))
	if length >= opt.MinLength && length <= opt.MaxLength {
		return urlPath
	}
	return ""
}

func validateInt(urlPath string, opt RouteOption) string {
	var numBytes []byte
	for _, char := range []byte(urlPath) {
		if isNumber(char) {
			numBytes = append(numBytes, char)
		} else {
			break
		}
		if uint8(len(numBytes)) >= opt.MaxLength {
			break
		}
	}
	if uint8(len(numBytes)) >= opt.MinLength {
		return string(numBytes)
	}
	return ""
}

var (
	wordReg, _ = regexp.Compile("^[\\w]+")
)

func validateWord(urlPath string, opt RouteOption) string {
	bytes := wordReg.Find([]byte(urlPath))
	if uint8(len(bytes)) >= opt.MinLength && uint8(len(bytes)) <= opt.MaxLength {
		return string(bytes)
	}
	return ""
}

func validateEnum(urlPath string, opt RouteOption) string {
	if len(opt.Setting) == 0 {
		return ""
	}
	var splits = strings.Split(opt.Setting, "|")
	for _, value := range splits {
		if strings.HasPrefix(urlPath, value) {
			return value
		}
	}
	return ""
}

var (
	acReg, _ = regexp.Compile("^[a-zA-Z0-9_]+(-[a-zA-Z0-9_]+)*")
)

func validateActionName(urlPath string, opt RouteOption) string {
	bytes := acReg.Find([]byte(urlPath))
	return string(bytes)
}
