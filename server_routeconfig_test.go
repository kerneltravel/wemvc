package wemvc

import "testing"

func Test_genFriendlyActionName(t *testing.T) {
	if genFriendlyActionName("GetIndex") != "get-index" {
		t.Fatal("Failed")
	}
	if genFriendlyActionName("GET_Index") != "get-index" {
		t.Fatal("Failed")
	}

	if genFriendlyActionName("GetUser___Info_") != "get-user-info" {
		t.Fatal("failed")
	}

	if genFriendlyActionName("AboutUs") != "about-us" {
		t.Fatal("failed")
	}
}

func Test_controllerInfo_findActionName(t *testing.T) {
	cInfo := &controllerInfo{
		Actions: map[string]string {
			"get-index": "GetIndex",
		},
	}
	println(cInfo.findActionName("GET", "index", true))
	if cInfo.findActionName("GET", "index", true) != "GetIndex" {
		t.Fatal("Test Failed")
	}
}