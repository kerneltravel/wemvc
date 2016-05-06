package controllers

import "github.com/Simbory/wemvc"

type Home struct {
	wemvc.Controller
}

func (this Home) GetIndex() wemvc.ActionResult {
	this.ViewData["msg"] = this.Session().Get("msg")
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	return this.View()
}

func (this Home) PostIndex() wemvc.ActionResult {
	msg := this.Request.Form.Get("msg")
	this.Session().Set("msg", msg)
	this.ViewData["msg"] = this.Session().Get("msg")
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	this.ViewData["s"] = this.Session().Get("s")
	return this.View()
}