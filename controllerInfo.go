package wemvc

import (
	"reflect"
	"strings"
)

type controllerInfo struct {
	controllerName string
	controllerType reflect.Type
	actions        map[string]string
}

func (cInfo *controllerInfo) containsAction(action string) bool {
	if cInfo == nil || cInfo.actions == nil {
		return false
	}
	name, ok := cInfo.actions[action]
	return ok && len(name) > 0
}

func createControllerInfo(t reflect.Type) *controllerInfo {
	typeName := t.String()
	if strings.HasPrefix(typeName, "*") {
		panic("Invalid controller type: \"" + typeName + "\"")
	}
	numMethod := t.NumMethod()
	if numMethod < 1 {
		panic("Thhe controller \"" + typeName + "\" has no action method")
	}
	obj := reflect.New(t)
	var methods []string
	for i := 0; i < numMethod; i++ {
		methodName := t.Method(i).Name
		method := obj.MethodByName(methodName)
		if !strings.HasSuffix(method.Type().String(), "wemvc.ActionResult") {
			continue
		}
		methods = append(methods, methodName)
	}
	if len(methods) < 1 {
		panic("The controller \"" + typeName + "\" has no action method")
	}
	actions := make(map[string]string)
	for _, m := range methods {
		actions[strings.ToLower(m)] = m
	}
	return &controllerInfo{
		controllerName: typeName,
		controllerType: t,
		actions:        actions,
	}
}
