package controllers

import "github.com/Simbory/wemvc"

type TestController struct  {
	wemvc.Controller
}

func (this TestController) GetIndex() wemvc.ActionResult {
	var msg string
	this.Session().Set("msg", "hello, world")
	tmpMsg := this.Session().Get("msg")
	if tmpMsg == nil {
		msg = ""
	} else {
		msg = tmpMsg.(string)
	}
	res := wemvc.NewActionResult()
	res.SetContentType("text/html")
	res.SetStatusCode(201)
	res.Write([]byte(msg))
	return res
}