package wemvc

import (
	"regexp"
	"strings"
)

var (
	acReg, _   = regexp.Compile("^[a-zA-Z0-9_]+(-[a-zA-Z0-9_]+)*")
	wordReg, _ = regexp.Compile("^[\\w]+")
)

// RouteFunc define the route check function
type RouteValidateFunc func(urlPath string, opt *RouteOption) string

func validateAny(urlPath string, opt *RouteOption) string {
	var length = uint8(len(urlPath))
	if length >= opt.MinLength && length <= opt.MaxLength {
		return urlPath
	}
	return ""
}

func validateInt(urlPath string, opt *RouteOption) string {
	var numBytes []byte
	for _, char := range str2Byte(urlPath) {
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
		return byte2Str(numBytes)
	}
	return ""
}

func validateWord(urlPath string, opt *RouteOption) string {
	bytes := wordReg.Find(str2Byte(urlPath))
	if uint8(len(bytes)) >= opt.MinLength && uint8(len(bytes)) <= opt.MaxLength {
		return byte2Str(bytes)
	}
	return ""
}

func validateEnum(urlPath string, opt *RouteOption) string {
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

func validateActionName(urlPath string, opt *RouteOption) string {
	bytes := acReg.Find(str2Byte(urlPath))
	if uint8(len(bytes)) > opt.MaxLength || uint8(len(bytes)) < opt.MinLength {
		return ""
	}
	return byte2Str(bytes)
}
