package wemvc

import (
	"os"
	"errors"
	"path/filepath"
	"fmt"
	"regexp"
	"bytes"
	"html/template"
)

type viewContainer struct {
	viewDir string
	views   map[string]*view
}

func (app *viewContainer)addView(name string, v *view) {
	if app.views == nil {
		app.views = make(map[string]*view)
	}
	app.views[name] = v
}

func (app *viewContainer)getView(name string) *view {
	v,ok := app.views[name]
	if !ok {
		return nil
	}
	return v
}

func (app *viewContainer)compileViews(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	vf := &viewFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return vf.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	for _, v := range vf.files {
		for _, file := range v {
			t, err := getTemplate(vf.root, file, v...)
			v := &view{tpl: t, err: err}
			app.addView(file, v)
		}
	}
	return nil
}

func (app *viewContainer)renderView(viewPath string, viewData interface{}) (template.HTML, int) {
	ext, _ := regexp.Compile(`\.[hH][tT][mM][lL]?$`)
	if !ext.MatchString(viewPath) {
		viewPath = viewPath + ".html"
	}

	tpl := app.getView(viewPath)
	if tpl == nil {
		return template.HTML("cannot find the view " + viewPath), 500
	}
	if tpl.err != nil {
		return template.HTML(tpl.err.Error()), 500
	}
	if tpl.tpl == nil {
		return template.HTML("cannot find the view " + viewPath), 500
	}
	var buf = &bytes.Buffer{}
	err := tpl.tpl.Execute(buf, viewData)
	if err != nil {
		return template.HTML(err.Error()), 500
	}
	result := template.HTML(buf.Bytes())
	return result, 200
}