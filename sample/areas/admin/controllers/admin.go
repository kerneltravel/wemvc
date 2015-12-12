package controllers

import "github.com/Simbory/wemvc"

type Admin struct {
	AuthController
}

func (this Admin) Get() wemvc.ActionResult {
	return this.Content("Hello, admin!")
}
