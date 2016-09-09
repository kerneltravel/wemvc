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
	server        *server
}

func (cInfo *controllerInfo) containsAction(action string) bool {
	if cInfo == nil || cInfo.Actions == nil {
		return false
	}
	name, ok := cInfo.Actions[action]
	return ok && len(name) > 0
}

func newControllerInfo(app *server, namespace string, t reflect.Type, defaultAction string) *controllerInfo {
	typeName := t.String()
	if strings.HasPrefix(typeName, "*") {
		panic(errInvalidCtrlType(typeName))
	}
	numMethod := t.NumMethod()
	if numMethod < 1 {
		panic(errCtrlNoAction(typeName))
	}
	obj := reflect.New(t)
	var methods []string
	app.logWriter().Println("Analyze controller", typeName)
	for i := 0; i < numMethod; i++ {
		methodName := t.Method(i).Name
		method := obj.MethodByName(methodName)
		methodType := method.Type().String()
		if !strings.HasSuffix(methodType, "wemvc.Result") && !strings.HasSuffix(methodType, "interface {}") {
			app.logWriter().Println("    Ignore method", methodName)
			continue
		} else {
			app.logWriter().Println("    Found action method", methodName)
		}
		methods = append(methods, methodName)
	}
	if len(methods) < 1 {
		panic(errCtrlNoAction(typeName))
	}
	actions := make(map[string]string)
	for _, m := range methods {
		actions[strings.ToLower(m)] = m
	}
	return &controllerInfo{
		NsName:        namespace,
		CtrlName:      typeName,
		CtrlType:      t,
		Actions:       actions,
		DefaultAction: defaultAction,
		server:        app,
	}
}
