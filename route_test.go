package wemvc

import (
	"testing"
	"reflect"
)

func Test_splitUrlPath(t *testing.T) {
	_,err := splitUrlPath("")
	if err == nil {
		t.Error("test 1 failed")
	}
	res2,err := splitUrlPath("/")
	if err != nil && res2 != nil {
		t.Error("test 2 failed")
	}
	_,err = splitUrlPath("/test/../nil")
	if err == nil {
		t.Error("test 3 failed")
	}
	res4,err := splitUrlPath("/test/./nil")
	if err != nil || len(res4) != 2 || res4[0] != "test" || res4[1] != "nil" {
		t.Error("test 4 failed")
	}
	res5,err := splitUrlPath("/test/{name}/nil/")
	if err != nil || len(res5) != 3 || res5[0] != "test" || res5[1] != "{name}" || res5[2] != "nil" {
		t.Error("test 5 failed")
	}
}

func Test_detectNodeType(t *testing.T) {
	if detectNodeType("/") != root {
		t.Error("test 1 failed")
	}
	if detectNodeType("test-a") != static {
		t.Error("test 2 failed")
	}
	if detectNodeType("edit-{user}") != param {
		t.Error("test 3 failed")
	}
}

type testCtrl struct {
	Controller
}
func (t testCtrl) Index() Result {
	return t.PlainText("test")
}

func newCtrlInfo() *controllerInfo {
	var app = newServer("C:\\www")
	var c = testCtrl{}
	return newControllerInfo(app, "", reflect.TypeOf(c), "Index")
}

func Test_newRouteDepth(t *testing.T) {
	var ctrlInfo = newCtrlInfo()
	routeDepth,err := newRouteBranch("/test/{action}", ctrlInfo)
	if err != nil {
		t.Error(err.Error())
	}
	if routeDepth.Path != "test" {
		t.Error("Depth Error")
	}
	if len(routeDepth.Children) != 1 || routeDepth.Children[0].Path != "{action}" {
		t.Error("Detect Child Error")
	}
}

func Test_newRootNode(t *testing.T) {
	var ctrlInfo = newCtrlInfo()
	var root = newRootNode()
	if err := root.addRoute("/", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/test/{action}", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/test/{fast}/nihao", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	println(string(data2Json(root)))
}

func Test_map(t *testing.T) {
	var mapData map[string]string
	var fun = func(m *map[string]string) {
		if *m == nil {
			*m = make(map[string]string)
		}
		(*m)["tets"] = "test1"
		(*m)["teta"] = "test2"
	}
	fun(&mapData)
	println(string(data2Json(mapData)))
}