package controllers

import "github.com/Simbory/wemvc"

type Edit struct {
	wemvc.Controller
}

func (c Edit) GetEdit() wemvc.Response {
	var action = c.RouteData.ByName("action")
	var id = c.RouteData.ByName("id")
	return c.Content(action + ":" + id)
}