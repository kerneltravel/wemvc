package controllers

import "github.com/Simbory/wemvc"

type TestController struct  {
	wemvc.Controller
}

func (this TestController) GetIndex() wemvc.ActionResult {
	var msg string
	tmpMsg := this.Session().Get("msg")
	if tmpMsg == nil {
		msg = ""
		this.Session().Set("msg", "hello, world")
	} else {
		msg = tmpMsg.(string)
	}
	res := wemvc.NewActionResult()
	res.SetContentType("text/plain")
	res.SetStatusCode(201)
	res.Write([]byte(msg))
	return res
}