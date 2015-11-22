package controllers

import "github.com/Simbory/wemvc"

type Admin struct {
	AuthController
}

func (this Admin) Get() wemvc.Response {
	return this.Content("Hello, admin!")
}
