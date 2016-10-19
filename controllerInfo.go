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

func (ctrlInfo *controllerInfo) containsAction(actionName, method string) (bool, string) {
	if len(ctrlInfo.Actions) == 0 || len(actionName) == 0 || len(method) == 0 {
		return false, ""
	}
	actionName = strings.Replace(strings.ToLower(actionName), "-", "_", -1)
	methodName, ok := ctrlInfo.Actions[method+actionName]
	if ok {
		return true, methodName
	}
	methodName, ok = ctrlInfo.Actions[method+"_"+actionName]
	if ok {
		return true, methodName
	}
	methodName, ok = ctrlInfo.Actions[actionName]
	if ok {
		return true, methodName
	}
	return false, ""
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
	//app.logWriter().Println("Analyze controller", typeName)
	for i := 0; i < numMethod; i++ {
		methodName := t.Method(i).Name
		method := obj.MethodByName(methodName)
		methodType := method.Type().String()
		if !strings.HasSuffix(methodType, "wemvc.Result") && !strings.HasSuffix(methodType, "interface {}") {
			//app.logWriter().Println("    Ignore method", methodName)
			continue
			//} else {
			//	app.logWriter().Println("    Found action method", methodName)
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
