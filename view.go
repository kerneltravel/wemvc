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
)

type viewFile struct {
	root  string
	files map[string][]string
}

type view struct {
	tpl *template.Template
	err error
}

var views map[string]*view

func addView(name string, v *view) {
	if views == nil {
		views = make(map[string]*view)
	}
	views[name] = v
}

func getView(name string) *view {
	if views != nil {
		return views[name]
	}
	return nil
}

func (self *viewFile) visit(paths string, f os.FileInfo, err error) error {
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
	a = a[len([]byte(self.root)):]
	file := strings.TrimLeft(replace.Replace(string(a)), "/")
	subdir := filepath.Dir(file)
	if _, ok := self.files[subdir]; ok {
		self.files[subdir] = append(self.files[subdir], file)
	} else {
		m := make([]string, 1)
		m[0] = file
		self.files[subdir] = m
	}
	return nil
}

func buildViews(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return errors.New("dir open err")
		}
	}
	self := &viewFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	for _, v := range self.files {
		for _, file := range v {
			t, err := getTemplate(self.root, file, v...)
			v := &view{tpl:t, err:err}
			addView(file, v)
		}
	}
	return nil
}

func getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	var submods [][]string
	t, submods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = getTplLoop(t, root, submods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileabspath string
	if filepath.HasPrefix(file, "../") {
		fileabspath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileabspath = filepath.Join(root, file)
	}
	if e := IsFile(fileabspath); !e {
		var msg = "can't find template file \"" + file + "\""
		return nil, [][]string{}, errors.New(msg)
	}
	data, err := ioutil.ReadFile(fileabspath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("{{" + "[ ]*template[ ]+\"([^\"]+)\"")
	allsub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allsub {
		if len(m) == 2 {
			tlook := t.Lookup(m[1])
			if tlook != nil {
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
	return t, allsub, nil
}

func getTplLoop(t0 *template.Template, root string, submods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range submods {
		if len(m) == 2 {
			templ := t.Lookup(m[1])
			if templ != nil {
				continue
			}
			//first check filename
			for _, otherfile := range others {
				if otherfile == m[1] {
					var submods1 [][]string
					t, submods1, err = getTplDeep(root, otherfile, "", t)
					if err != nil {
						return nil, err
					} else if submods1 != nil && len(submods1) > 0 {
						t, err = getTplLoop(t, root, submods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherfile := range others {
				fileabspath := filepath.Join(root, otherfile)
				data, err := ioutil.ReadFile(fileabspath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allsub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allsub {
					if len(sub) == 2 && sub[1] == m[1] {
						var submods1 [][]string
						t, submods1, err = getTplDeep(root, otherfile, "", t)
						if err != nil {
							return nil, err
						} else if submods1 != nil && len(submods1) > 0 {
							t, err = getTplLoop(t, root, submods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

func renderView(viewPath string, viewData interface{}) (template.HTML, int) {
	ext, _ := regexp.Compile(`\.[hH][tT][mM][lL]?$`)
	if !ext.MatchString(viewPath) {
		viewPath = viewPath + ".html"
	}

	tpl := getView(viewPath)
	if tpl == nil {
		return template.HTML("cannot find the view " + viewPath), 500
	}
	if tpl.err != nil {
		return template.HTML(tpl.err.Error()),500
	}
	if tpl.tpl == nil {
		return template.HTML("cannot find the view " + viewPath),500
	}
	var buf = &bytes.Buffer{}
	err := tpl.tpl.Execute(buf, viewData)
	if err != nil {
		return template.HTML(err.Error()), 500
	}
	result := template.HTML(buf.Bytes())
	return result, 200
}
