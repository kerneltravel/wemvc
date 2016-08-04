package wemvc

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
)

type viewContainer struct {
	viewDir string
	views   map[string]*view
}

func (vc *viewContainer) addView(name string, v *view) {
	if vc.views == nil {
		vc.views = make(map[string]*view)
	}
	vc.views[name] = v
}

func (vc *viewContainer) getView(name string) *view {
	v, ok := vc.views[name]
	if !ok {
		return nil
	}
	return v
}

func (vc *viewContainer) compileViews(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	app.logWriter().Println("compile view files in dir", dir)
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
			vc.addView(file, v)
		}
	}
	return nil
}

func (vc *viewContainer) renderView(viewPath string, viewData interface{}) (template.HTML, int) {
	ext, _ := regexp.Compile(`\.[hH][tT][mM][lL]?$`)
	if !ext.MatchString(viewPath) {
		viewPath = viewPath + ".html"
	}

	tpl := vc.getView(viewPath)
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
