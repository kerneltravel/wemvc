package wemvc

import (
	"strings"
	"reflect"
)

type routeConfig struct {
	name      string
	namespace string
	routePath string
	c         interface{}
	action    string
}

func genFriendlyActionName(methodName string) string {
	var wordArr []byte
	for i := 0; i < len(methodName); i++ {
		char := methodName[i]
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			wordArr = append(wordArr, char)
		} else {
			if len(wordArr) > 0 {
				lastChar := wordArr[len(wordArr) - 1]
				if lastChar != '-' && i < len(methodName) - 1 {
					wordArr = append(wordArr, '-')
				}
			}
			if char >= 'A' && char <= 'Z' {
				wordArr = append(wordArr, byte(char+32))
			}
		}
	}
	return byte2Str(wordArr)
}

func (r *routeConfig) genCtrlInfo(friendlyAction bool) *controllerInfo {
	t := reflect.TypeOf(r.c)
	typeName := t.String()
	if strings.HasPrefix(typeName, "*") {
		panic(errInvalidCtrlType(typeName))
	}
	numMethod := t.NumMethod()
	if numMethod < 1 {
		panic(errCtrlNoAction(typeName))
	}
	methods := make([]string, 0, numMethod)
	for i := 0; i < numMethod; i++ {
		methodInfo := t.Method(i)
		numIn := methodInfo.Type.NumIn()
		numOut := methodInfo.Type.NumOut()
		if numIn != 1 || numOut != 1 {
			continue
		}
		methodName := methodInfo.Name
		methods = append(methods, methodName)
	}
	if len(methods) < 1 {
		panic(errCtrlNoAction(typeName))
	}
	actions := make(map[string]string, len(methods))
	for _, m := range methods {
		var actionName string
		if friendlyAction {
			actionName = genFriendlyActionName(m)
		} else {
			actionName = m
		}
		actions[actionName] = m
	}
	return &controllerInfo{
		NsName:        r.namespace,
		CtrlName:      getControllerName(t),
		CtrlType:      t,
		Actions:       actions,
		DefaultAction: r.action,
	}
}