package wemvc

import "strings"

// RouteFunc define the route check function
type RouteFunc func(urlPath string, opt RouteOption) string

func stringCheck(urlPath string, opt RouteOption) string {
	var length = uint8(len(urlPath))
	if length >= opt.MinLength && length <= opt.MaxLength {
		return urlPath
	}
	return ""
}

func intCheck(urlPath string, opt RouteOption) string {
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

//noinspection GoUnusedParameter
func wordCheck(urlPath string, opt RouteOption) string {
	var bytes []byte
	for i := 0; i < len(urlPath); i++ {
		if isWord(urlPath[i]) || isNumber(urlPath[i]) {
			bytes = append(bytes, urlPath[i])
		} else {
			break
		}
	}
	return string(bytes)
}

func enumCheck(urlPath string, opt RouteOption) string {
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
