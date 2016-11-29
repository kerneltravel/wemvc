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

func (ctrlInfo *controllerInfo) containsAction(actionName, method string) (bool, string) {
	if len(ctrlInfo.Actions) == 0 || len(actionName) == 0 || len(method) == 0 {
		return false, ""
	}
	actionName = strings.Replace(strings.ToLower(actionName), "-", "_", -1)
	methodName, ok := ctrlInfo.Actions[strAdd(method, actionName)]
	if ok {
		return true, methodName
	}
	methodName, ok = ctrlInfo.Actions[strAdd(method, "_", actionName)]
	if ok {
		return true, methodName
	}
	methodName, ok = ctrlInfo.Actions[actionName]
	if ok {
		return true, methodName
	}
	return false, ""
}

func newControllerInfo(namespace string, t reflect.Type, defaultAction string) *controllerInfo {
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
		actions[strings.ToLower(m)] = m
	}
	return &controllerInfo{
		NsName:        namespace,
		CtrlName:      getControllerName(t),
		CtrlType:      t,
		Actions:       actions,
		DefaultAction: defaultAction,
	}
}
