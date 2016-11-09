package wemvc

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"io/ioutil"
	"regexp"
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

func (vc *viewContainer) getTemplate(file, viewExt string, funcMap template.FuncMap, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	var subMods [][]string
	t, subMods, err = vc.getTemplateDeep(file, viewExt, "", t)
	if err != nil {
		return nil, err
	}
	t, err = vc.getTemplateLoop(t, viewExt, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func (vc *viewContainer) getTemplateDeep(file, viewExt, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if filepath.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(vc.viewDir, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(vc.viewDir, file)
	}
	if e := IsFile(fileAbsPath); !e {
		return nil, [][]string{}, errNotFoundTpl(file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("{{" + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			look := t.Lookup(m[1])
			if look != nil {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(m[1]), viewExt) {
				continue
			}
			t, _, err = vc.getTemplateDeep(m[1], viewExt, file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func (vc *viewContainer) getTemplateLoop(t0 *template.Template, viewExt string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = vc.getTemplateDeep(otherFile, viewExt, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = vc.getTemplateLoop(t, viewExt, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(vc.viewDir, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = vc.getTemplateDeep(otherFile, viewExt, "", t)
						if err != nil {
							return nil, err
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = vc.getTemplateLoop(t, viewExt, subMods1, others...)
						}
						break
					}
				}
			}
		}
	}
	return
}

func (vc *viewContainer) compileViews(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return errOpenDir
	}
	vc.viewDir = dir
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
			t, err := vc.getTemplate(file, vf.viewExt, vc.funcMaps, v...)
			v := &view{tpl: t, err: err}
			vc.addView(file, v)
		}
	}
	return nil
}

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
