package wemvc

import (
	"reflect"
	"strings"
)

type controllerInfo struct {
	NsName        string
	CtrlName      string
	CtrlType      reflect.Type
	Actions       map[string]string
	DefaultAction string
}

func (ctrlInfo *controllerInfo) findActionName(actionName, method string, friendly bool) string {
	if len(actionName) == 0 || len(method) == 0 {
		return ""
	}
	if friendly {
		methodName,ok := ctrlInfo.Actions[strings.ToLower(strAdd(method, "-", actionName))]
		if ok {
			return methodName
		}
		return ctrlInfo.Actions[strings.ToLower(actionName)]
	} else {
		methodName,ok := ctrlInfo.Actions[strings.ToLower(strAdd(method, actionName))]
		if ok {
			return methodName
		}
		methodName,ok = ctrlInfo.Actions[strings.ToLower(strAdd(method, "_", actionName))]
		if ok {
			return methodName
		}
		return ctrlInfo.Actions[strings.ToLower(actionName)]
	}
}
