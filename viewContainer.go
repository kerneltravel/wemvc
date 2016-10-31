package wemvc

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type viewContainer struct {
	viewExt  string
	viewDir  string
	views    map[string]*view
	funcMaps template.FuncMap
}

func (vc *viewContainer) addViewFunc(name string, f interface{}) {
	if len(name) < 1 || f == nil {
		return
	}
	if vc.funcMaps == nil {
		vc.funcMaps = make(template.FuncMap)
	}
	vc.funcMaps[name] = f
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
		return errOpenDir
	}
	vf := &viewFile{
		root:    dir,
		files:   make(map[string][]string),
		viewExt: vc.viewExt,
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return vf.visit(path, f, err)
	})
	if err != nil {
		return err
	}
	for _, v := range vf.files {
		for _, file := range v {
			t, err := getTemplate(vf.root, file, vf.viewExt, vc.funcMaps, v...)
			v := &view{tpl: t, err: err}
			vc.addView(file, v)
		}
	}
	return nil
}

// renderView render the view template with ViewData and get the result
func (vc *viewContainer) renderView(viewPath string, viewData interface{}) ([]byte, error) {
	if len(viewPath) < 1 {
		return nil, errEmptyViewPath
	}
	if !strings.HasSuffix(viewPath, vc.viewExt) {
		viewPath = viewPath + vc.viewExt
	}
	tpl := vc.getView(viewPath)
	if tpl == nil {
		return nil, errViewPathNotFound(viewPath)
	}
	if tpl.err != nil {
		return nil, tpl.err
	}
	if tpl.tpl == nil {
		return nil, errViewPathNotFound(viewPath)
	}
	buf := &bytes.Buffer{}
	err := tpl.tpl.Execute(buf, viewData)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
