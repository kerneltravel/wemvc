package controllers
import "github.com/Simbory/wemvc"

type Download struct {
	wemvc.Controller
}

func (t Download)Get() wemvc.Response {
	var file = t.Request.URL.Query().Get("file")
	if len(file) < 1 {
		return t.NotFound()
	}
	file = wemvc.App.MapPath(file)
	if  !wemvc.IsFile(file) {
		return t.NotFound()
	}
	return t.File(file, "text/xml")
}