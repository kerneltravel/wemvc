package controllers

import "github.com/Simbory/wemvc"

type Home struct {
	wemvc.Controller
	name string
}

func (this Home)Get() wemvc.Response {
	if (this.name == "simbory") {
		panic("found framework bug")
	}
	this.name = "simbory"
	this.ViewData["msg"] = wemvc.App.GetConfig().GetSetting("isDebug")
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	return this.View("home")
}