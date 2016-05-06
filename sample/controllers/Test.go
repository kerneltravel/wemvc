package controllers

import "github.com/Simbory/wemvc"

type TestController struct  {
}

func (this TestController) GetIndex() wemvc.ActionResult {
	res := wemvc.NewActionResult()
	res.SetContentType("text/html")
	res.SetStatusCode(201)
	res.Write([]byte("test App"))
	return res
}