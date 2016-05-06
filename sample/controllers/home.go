package controllers

import "github.com/Simbory/wemvc"

type Home struct {
	wemvc.Controller
}

func (this Home) GetIndex() wemvc.ActionResult {
	this.ViewData["msg"] = "this is get action result"
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	return this.View()
}

func (this Home) Index() wemvc.ActionResult {
	this.ViewData["msg"] = wemvc.App.GetConfig().GetSetting("isDebug")
	this.ViewData["wwwroot"] = wemvc.App.GetWebRoot()
	return this.View()
}