package wemvc

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/Simbory/wemvc/utils"
)

type viewFile struct {
	root  string
	files map[string][]string
}

type view struct {
	tpl *template.Template
	err error
}

func (app *server)addView(name string, v *view) {
	app.views[name] = v
}

func (app *server)getView(name string) *view {
	v,ok := app.views[name]
	if !ok {
		return nil
	}
	return v
}

func (vf *viewFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(paths), ".html") {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	a := []byte(paths)
	a = a[len([]byte(vf.root)):]
	file := strings.TrimLeft(replace.Replace(string(a)), "/")
	subDir := filepath.Dir(file)
	if _, ok := vf.files[subDir]; ok {
		vf.files[subDir] = append(vf.files[subDir], file)
	} else {
		m := make([]string, 1)
		m[0] = file
		vf.files[subDir] = m
	}
	return nil
}

func (app *server)buildViews(dir string) error {
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

func getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = getTplLoop(t, root, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if filepath.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(root, file)
	}
	if e := utils.IsFile(fileAbsPath); !e {
		var msg = "can't find template file \"" + file + "\""
		return nil, [][]string{}, errors.New(msg)
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
			if !strings.HasSuffix(strings.ToLower(m[1]), ".html") {
				continue
			}
			t, _, err = getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func getTplLoop(t0 *template.Template, root string, subMods [][]string, others ...string) (t *template.Template, err error) {
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
					t, subMods1, err = getTplDeep(root, otherFile, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = getTplLoop(t, root, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(root, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = getTplDeep(root, otherFile, "", t)
						if err != nil {
							return nil, err
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = getTplLoop(t, root, subMods1, others...)
						}
						break
					}
				}
			}
		}
	}
	return
}

func (app *server)renderView(viewPath string, viewData interface{}) (template.HTML, int) {
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
