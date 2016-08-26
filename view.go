package wemvc

import (
	"errors"
	"github.com/Simbory/wemvc/utils"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type view struct {
	tpl *template.Template
	err error
}

type viewFile struct {
	viewExt string
	root    string
	files   map[string][]string
}

func (vf *viewFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(paths), vf.viewExt) {
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

func getTemplate(root, file, viewExt string, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, file, viewExt, "", t)
	if err != nil {
		return nil, err
	}
	t, err = getTplLoop(t, root, viewExt, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func getTplDeep(root, file, viewExt, parent string, t *template.Template) (*template.Template, [][]string, error) {
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
			if !strings.HasSuffix(strings.ToLower(m[1]), viewExt) {
				continue
			}
			t, _, err = getTplDeep(root, m[1], viewExt, file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func getTplLoop(t0 *template.Template, root, viewExt string, subMods [][]string, others ...string) (t *template.Template, err error) {
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
					t, subMods1, err = getTplDeep(root, otherFile, viewExt, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = getTplLoop(t, root, viewExt, subMods1, others...)
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
						t, subMods1, err = getTplDeep(root, otherFile, viewExt, "", t)
						if err != nil {
							return nil, err
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = getTplLoop(t, root, viewExt, subMods1, others...)
						}
						break
					}
				}
			}
		}
	}
	return
}
