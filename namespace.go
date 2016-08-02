package wemvc

import (
	"strings"
	"github.com/Simbory/wemvc/utils"
	"os"
	"path/filepath"
	"regexp"
	"bytes"
	"html/template"
)

type NamespaceSection interface {
	GetName() string
	GetNsDir() string
	Route(string, interface{}, ...string) NamespaceSection
	GetSetting(string) string
}

type nsSettingGroup struct {
	Settings     []setting `xml:"add"`
}

type namespace struct {
	name     string
	views    map[string]*view
	server   *server
	settings map[string]string
}

func (ns *namespace)GetName() string {
	return ns.name
}

func (ns *namespace)GetNsDir() string {
	return ns.server.mapPath(ns.GetName())
}

func (ns *namespace)Route(routePath string, c interface{}, defaultAction ...string) NamespaceSection {
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}
	var nsName = ns.GetName()
	routePath = nsName + routePath
	var action = "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	ns.server.route(nsName, routePath, c, action)
	return ns
}

func (ns *namespace)GetSetting(key string) string {
	v,ok := ns.settings[key]
	if ok {
		return v
	}
	return ""
}

func (ns *namespace)nsSettingFile() string {
	return ns.server.mapPath(ns.GetName() + "/settings.config")
}

func (ns *namespace)isConfigFile(f string) bool {
	return ns.nsSettingFile() == f
}

func (ns *namespace)nsViewDir() string {
	return ns.server.mapPath(ns.GetName() + "/views")
}

func (ns *namespace) isInViewFolder(f string) bool {
	var viewPath = ns.nsViewDir()
	return strings.HasPrefix(f, viewPath)
}

func (ns *namespace)loadConfig() {
	var path = ns.nsSettingFile()
	if utils.IsFile(path) {
		var settings = &nsSettingGroup{}
		ns.server.logWriter().Println("load config file '"+ path + "' for namespace '"+ ns.GetName() + "'")
		err := utils.File2Xml(path, settings)
		if err != nil {
			return
		}
		settingMap := make(map[string]string)
		for _,s := range settings.Settings {
			settingMap[s.Key] = s.Value
		}
		ns.settings = settingMap
	}
}

func (ns *namespace)addView(name string, v *view) {
	if ns.views == nil {
		ns.views = make(map[string]*view)
	}
	ns.views[name] = v
}

func (ns *namespace)getView(name string) *view {
	if ns.views == nil {
		return nil
	}
	v,ok := ns.views[name]
	if !ok {
		return nil
	}
	return v
}

func (ns *namespace)compileViews() {
	ns.views = nil
	var dir = ns.nsViewDir()
	if utils.IsDir(dir) {
		vf := &viewFile{
			root:  dir,
			files: make(map[string][]string),
		}
		err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
			return vf.visit(path, f, err)
		})
		if err != nil {
			ns.server.logWriter().Printf("filepath.Walk() returned %v\n", err)
			return
		}
		for _, v := range vf.files {
			for _, file := range v {
				t, err := getTemplate(vf.root, file, v...)
				v := &view{tpl: t, err: err}
				ns.addView(file, v)
			}
		}
	}
}

func (ns *namespace)renderView(viewPath string, viewData interface{}) (template.HTML, int) {
	ext, _ := regexp.Compile(`\.[hH][tT][mM][lL]?$`)
	if !ext.MatchString(viewPath) {
		viewPath = viewPath + ".html"
	}

	tpl := ns.getView(viewPath)
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