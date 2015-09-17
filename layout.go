package wemvc

import (
	"html/template"
	"regexp"
	"strings"
	"path/filepath"
	"bytes"
)

func layout(viewPath string, viewData interface{}) template.HTML {
	ext,_ := regexp.Compile(`\.[hH][tT][mM][lL]?$`)
	if !ext.MatchString(viewPath) {
		viewPath = viewPath + ".html"
	}

	if !strings.HasPrefix(viewPath, "/") {
		viewPath = "/views/" + viewPath
	}
	var viewFile = App.MapPath(viewPath)
	println(viewFile)
	if !isFile(viewFile) {
		panic("Cannot find the view \"" + viewPath + "\"")
	}
	var filename = filepath.Base(viewFile)
	funcMap := template.FuncMap{
		"layout": layout,
	}
	tpl := template.New(filename).Funcs(funcMap)

	tpl,err := tpl.ParseFiles(viewFile, App.MapPath("/views/_layout.html"))
	if err != nil {
		panic(err)
	}
	var buf = &bytes.Buffer{}
	err = tpl.Execute(buf, viewData)
	result := template.HTML(buf.Bytes())
	println(result)
	return result
}