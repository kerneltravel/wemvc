package wemvc

import (
	"reflect"
	"testing"
)

func Test_splitUrlPath(t *testing.T) {
	_, err := splitURLPath("")
	if err == nil {
		t.Error("test 1 failed")
	}
	res2, err := splitURLPath("/")
	if err != nil && res2 != nil {
		t.Error("test 2 failed")
	}
	_, err = splitURLPath("/test/../nil")
	if err == nil {
		t.Error("test 3 failed")
	}
	res4, err := splitURLPath("/test/./nil")
	if err != nil || len(res4) != 2 || res4[0] != "test" || res4[1] != "nil" {
		t.Error("test 4 failed")
	}
	res5, err := splitURLPath("/test/<name>/nil/")
	if err != nil || len(res5) != 3 || res5[0] != "test" || res5[1] != "<name>" || res5[2] != "nil" {
		t.Error("test 5 failed")
	}
}

func Test_detectNodeType(t *testing.T) {
	if detectNodeType("/") != rtRoot {
		t.Error("test 1 failed")
	}
	if detectNodeType("test-a") != rtStatic {
		t.Error("test 2 failed")
	}
	if detectNodeType("edit-<user>") != rtParam {
		t.Error("test 3 failed")
	}
}

type testCtrl struct {
	Controller
}

func (t testCtrl) Index() ContentResult {
	return t.PlainText("test")
}

func newCtrlInfo() *controllerInfo {
	var app = newServer("C:\\www")
	var c = testCtrl{}
	return newControllerInfo(app, "", reflect.TypeOf(c), "Index")
}

func Test_newRouteDepth(t *testing.T) {
	var ctrlInfo = newCtrlInfo()
	b1, err := newRouteNode("/test/{action}", ctrlInfo)
	if err != nil {
		t.Error(err.Error())
	}
	if b1.Path != "test" {
		t.Error("Depth Error")
	}
	if len(b1.Children) != 1 || b1.Children[0].Path != "{action}" {
		t.Error("Detect Child Error")
	}
	_, err = newRouteNode("/*pathInfo/test", ctrlInfo)
	if err == nil {
		t.Error("Error check failed")
	} else {
		println(err.Error())
	}
	b2, err := newRouteNode("/test/*path", ctrlInfo)
	if err == nil {
		t.Error("Error check failed")
		println(string(data2Json(b2)))
	} else {
		println(err.Error())
	}
}

func Test_newRootNode(t *testing.T) {
	var ctrlInfo = newCtrlInfo()
	var root = newRouteTree()
	if err := root.addRoute("/", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/test", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/test/<action>", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/test/<year>/hello", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/<fast>/*pathInfo", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	if err := root.addRoute("/edit/<user:word(4)>/<action=edit-user>", ctrlInfo); err != nil {
		t.Error(err.Error())
	}
	println(string(data2Json(root)))
}

func Test_splitRouteParam(t *testing.T) {
	var path = "<word>-fag-<username><email:word(4)>"
	var splits = splitRouteParam(path)
	println(string(data2Json(splits)))
	if len(splits) != 4 || splits[0] != "<word>" || splits[1] != "-fag-" || splits[2] != "<username>" || splits[3] != "<email:word(4)>" {
		t.Error("test 1 failed")
	}
	println(string(data2Json(splitRouteParam("fdsgdfsg>fa<-fag>-"))))
}

func Test_checkParamName(t *testing.T) {
	if checkParamName("usename") == false {
		t.Error("Test 1 failed")
	}
	if checkParamName("usename_") == false {
		t.Error("Test 2 failed")
	}
	if checkParamName("use_name") == false {
		t.Error("Test 3 failed")
	}
	if checkParamName("usename1") == false {
		t.Error("Test 4 failed")
	}
	if checkParamName("usename 1") == true {
		t.Error("Test 5 failed")
	}
	if checkParamName("_usename1") == true {
		t.Error("Test 6 failed")
	}
	if checkParamName("123") == true {
		t.Error("Test 7 failed")
	}
	if checkParamName("_1") == true {
		t.Error("Test 8 failed")
	}
	if checkParamName("usename1+") == true {
		t.Error("Test 9 failed")
	}
	if checkParamName("1usename") == true {
		t.Error("Test 10 failed")
	}
}

func Test_checkParamOption(t *testing.T) {
	if checkParamOption("a2z(3~4)") == false {
		t.Error("test 1 failed")
	}
	if checkParamOption("value(aa|bb|cc)") == false {
		t.Error("test 1 failed")
	}
	if checkParamOption("word(45)") == false {
		t.Error("test 1 failed")
	}
}
