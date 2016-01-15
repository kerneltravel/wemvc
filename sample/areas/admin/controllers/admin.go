package controllers

import "github.com/Simbory/wemvc"

type Admin struct {
	wemvc.Controller
}

func (this Admin) Get() wemvc.ActionResult {
	return this.Content("Hello," + this.Items["name"].(string))
}
