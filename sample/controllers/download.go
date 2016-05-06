package controllers

import (
	"github.com/Simbory/wemvc"
	"github.com/Simbory/wemvc/utils"
)

type Download struct {
	wemvc.Controller
}

func (t Download)Get() wemvc.ActionResult {
	var file = t.Request.URL.Query().Get("file")
	if len(file) < 1 {
		return t.NotFound()
	}
	file = wemvc.App.MapPath(file)
	if  !utils.IsFile(file) {
		return t.NotFound()
	}
	return t.File(file, "text/xml")
}