package wemvc

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func isA2Z(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isWord(c byte) bool {
	return isA2Z(c) || c == '_'
}

func isNumber(c byte) bool {
	return (c >= '0' && c <= '9')
}

func splitURLPath(urlPath string) ([]string, error) {
	if len(urlPath) == 0 {
		return nil, errors.New("The URL path is empty")
	}
	p := strings.Trim(urlPath, "/")
	splits := strings.Split(p, "/")
	var result []string
	for _, s := range splits {
		if len(s) == 0 || s == "." {
			continue
		}
		if s == ".." {
			return nil, errors.New("Invalid URL path. The URL path cannot contains '..'")
		}
		result = append(result, s)
	}
	return result, nil
}

func detectNodeType(p string) pathType {
	if p == "/" {
		return rtRoot
	}
	if strings.Contains(p, string([]byte{rtParamBegin})) || strings.Contains(p, string([]byte{rtParamEnd})) {
		return rtParam
	}
	if p == rtPathInfo {
		return rtCatchAll
	}
	return rtStatic
}

func checkRoutePath(path string) error {
	var routeParams []string
	var paramChars []byte
	var inParamChar = false

	for i := 0; i < len(path); i++ {
		// param begin
		if path[i] == rtParamBegin {
			if len(paramChars) == 0 {
				inParamChar = true
				continue
			} else {
				return fmt.Errorf("the route param has no closing character '>': %d", i)
			}
		}
		// param end
		if path[i] == rtParamEnd {
			// check and ensure current route param is not empty
			if len(paramChars) == 0 {
				return fmt.Errorf("Invalid route parameter '<>' or the route parameter has no begining tag '<': %d", i)
			}
			curParam := strings.Split(string(paramChars), ":")[0]
			for _, tmp := range routeParams {
				if tmp == curParam {
					return fmt.Errorf("Duplicate route param '%s': %d", curParam, i)
				}
			}
			routeParams = append(routeParams, curParam)
			paramChars = make([]byte, 0)
			inParamChar = false
			continue
		}
		if inParamChar {
			if len(paramChars) == 0 {
				if isA2Z(path[i]) {
					paramChars = append(paramChars, path[i])
				} else {
					return fmt.Errorf("Invalid character '%c' at the beginin of the route param: %d", path[i], i)
				}
			} else {
				paramChars = append(paramChars, path[i])
			}
		}
	}
	if len(routeParams) > 255 {
		return errors.New("Too many route params: the maximum number of the route param is 255")
	}
	return nil
}

func splitRouteParam(path string) []string {
	var splits []string
	var byteQueue []byte
	for _, char := range []byte(path) {
		if char == rtParamEnd {
			byteQueue = append(byteQueue, char)
			if len(byteQueue) > 0 {
				splits = append(splits, string(byteQueue))
				byteQueue = nil
			}
		} else {
			if char == rtParamBegin && len(byteQueue) > 0 {
				splits = append(splits, string(byteQueue))
				byteQueue = nil
			}
			byteQueue = append(byteQueue, char)
		}
	}
	if len(byteQueue) > 0 {
		splits = append(splits, string(byteQueue))
	}
	return splits
}

func checkParamName(name string) bool {
	reg, _ := regexp.Compile("^[a-zA-Z][\\w]*$")
	return reg.Match([]byte(name))
}

func checkParamOption(optionStr string) bool {
	reg, _ := regexp.Compile("^[a-zA-Z][\\w]*\\(.+\\)$")
	return reg.Match([]byte(optionStr))
}

func checkNumber(opt string) bool {
	reg, _ := regexp.Compile("^[0-9]+$")
	return reg.Match([]byte(opt))
}

func checkNumberRange(optStr string) bool {
	reg, _ := regexp.Compile("^[0-9]+(~)+[0-9]+$")
	return reg.Match([]byte(optStr))
}

func analyzeParamOption(path string) ([]string, map[string]*RouteOption, error) {
	splitParams := splitRouteParam(path)
	optionMap := make(map[string]*RouteOption)
	var paramPath []string
	for _, sp := range splitParams {
		if strings.HasSuffix(sp, rtParamEndStr) && strings.HasPrefix(sp, rtParamBeginStr) {
			paramStr := strings.Trim(sp, rtParamBeginStr + rtParamEndStr)
			splits := strings.Split(paramStr, ":")
			// paramName: the name of the route param (with default value), like 'name', 'name=Steve Jobs' or 'name='
			paramName := splits[0]
			// paramOptionStr: the route param option
			paramOptionStr := ""
			if len(splits) == 1 {
				if paramName == "action" {
					paramOptionStr = "action"
				} else {
					paramOptionStr = "any"
				}
			}
			if len(splits) == 2 {
				paramOptionStr = splits[1]
				if len(paramOptionStr) == 0 {
					if paramName == "action" {
						paramOptionStr = "action"
					} else {
						paramOptionStr = "any"
					}
				}
			} else if len(splits) > 2 {
				return nil, nil, errors.New("Invalid route parameter setting: " + sp)
			}
			opt := RouteOption{}
			var eqIndex = strings.Index(paramName, "=")
			if eqIndex > 0 {
				defaultValue := paramName[eqIndex+1:]
				paramName = paramName[0:eqIndex]
				opt.DefaultValue = defaultValue
				opt.HasDefaultValue = true
			} else if !checkParamName(paramName) {
				return nil, nil, errors.New("Invalid route parameter name '" + paramName + "': " + sp)
			} else {
				opt.HasDefaultValue = false
			}
			if checkParamName(paramOptionStr) {
				opt.Validation = paramOptionStr
				opt.MaxLength = 255
				opt.MinLength = 1
			} else if checkParamOption(paramOptionStr) {
				optSplits := strings.Split(paramOptionStr, "(")
				if len(optSplits) != 2 {
					return nil, nil, errors.New("Invalid route parameter setting: " + sp)
				}
				opt.Validation = optSplits[0]
				var setting = strings.TrimRight(optSplits[1], ")")
				if strings.Contains(setting, ")") {
					return nil, nil, errors.New("Invalid route parameter setting: " + sp)
				}
				if checkNumber(setting) {
					i, err := strconv.ParseUint(setting, 10, 0)
					if err != nil {
						return nil, nil, err
					}
					opt.MaxLength = uint8(i)
					opt.MinLength = uint8(i)
				} else if checkNumberRange(setting) {
					numbers := strings.Split(setting, "~")
					min, err := strconv.ParseUint(numbers[0], 10, 0)
					if err != nil {
						return nil, nil, err
					}
					max, err := strconv.ParseUint(numbers[1], 10, 0)
					if err != nil {
						return nil, nil, err
					}
					if min < max {
						opt.MinLength = uint8(min)
						opt.MaxLength = uint8(max)
					} else {
						opt.MinLength = uint8(max)
						opt.MaxLength = uint8(min)
					}
				} else {
					opt.MaxLength = 255
					opt.MinLength = 1
					opt.Setting = setting
				}
			}
			optionMap[paramName] = &opt
			paramPath = append(paramPath, rtParamBeginStr +paramName+ rtParamEndStr)
		} else {
			paramPath = append(paramPath, sp)
		}
	}
	return paramPath, optionMap, nil
}
