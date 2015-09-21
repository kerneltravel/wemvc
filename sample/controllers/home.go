package controllers

import "github.com/Simbory/wemvc"

type Home struct {
	wemvc.Controller
}

func (this *Home)Get() wemvc.Response {
	this.ViewData["msg"] = wemvc.App.GetConfig().GetSetting("isDebug")
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	return this.View("home")
}