package wemvc

import (
	"reflect"
	"strings"
)

type controllerInfo struct {
	namespace      string
	controllerName string
	controllerType reflect.Type
	actions        map[string]string
	defaultAction  string
}

func (cInfo *controllerInfo) containsAction(action string) bool {
	if cInfo == nil || cInfo.actions == nil {
		return false
	}
	name, ok := cInfo.actions[action]
	return ok && len(name) > 0
}

func newControllerInfo(namespace string, t reflect.Type, defaultAction string) *controllerInfo {
	typeName := t.String()
	if strings.HasPrefix(typeName, "*") {
		panic(invalidCtrlTypeError(typeName))
	}
	numMethod := t.NumMethod()
	if numMethod < 1 {
		panic(ctrlNoActionError(typeName))
	}
	obj := reflect.New(t)
	var methods []string
	app.logWriter().Println("Analyze controller", typeName)
	for i := 0; i < numMethod; i++ {
		methodName := t.Method(i).Name
		method := obj.MethodByName(methodName)
		methodType := method.Type().String()
		if !strings.HasSuffix(methodType, "wemvc.ActionResult") && !strings.HasSuffix(methodType, "interface {}") {
			app.logWriter().Println("    Ignore method", methodName)
			continue
		} else {
			app.logWriter().Println("    Found action method", methodName)
		}
		methods = append(methods, methodName)
	}
	if len(methods) < 1 {
		panic(ctrlNoActionError(typeName))
	}
	actions := make(map[string]string)
	for _, m := range methods {
		actions[strings.ToLower(m)] = m
	}
	return &controllerInfo{
		namespace:      namespace,
		controllerName: typeName,
		controllerType: t,
		actions:        actions,
		defaultAction:  defaultAction,
	}
}
