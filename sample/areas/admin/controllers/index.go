package controllers
import "github.com/Simbory/wemvc"

type Index struct {
	AdminController
}

func (this Index)Get() wemvc.Response {
	return this.Content("Hello, admin!")
}