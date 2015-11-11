package controllers
import "github.com/Simbory/wemvc"

type Index struct {
	wemvc.Controller
}

func (this *Index)Get() wemvc.Response {
	return this.Content("Hello, admin!")
}