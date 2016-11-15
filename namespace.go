package wemvc

import (
	"strings"
	"runtime"
)

// NamespaceSection the namespace section interface
type NsSection interface {
	GetName() string
	GetNsDir() string
	Route(string, interface{}, ...string) NsSection
	Filter(string, CtxFilter) NsSection
	StaticDir(string) NsSection
	StaticFile(string) NsSection
	GetSetting(string) string
	AddViewFunc(name string, f interface{}) NsSection
}

type nsSettingGroup struct {
	Settings []struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	} `xml:"add"`
}

type namespace struct {
	name     string
	server   *server
	settings map[string]string
	viewContainer
	filterContainer
}

func (ns *namespace) Name() string {
	return ns.name
}

func (ns *namespace) Dir() string {
	return ns.server.mapPath(ns.Name())
}

func (ns *namespace) Route(routePath string, c interface{}, defaultAction ...string) NsSection {
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}
	var nsName = ns.Name()
	routePath = nsName + routePath
	var action = "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	ns.server.addRoute(nsName, routePath, c, action)
	return ns
}

func (ns *namespace) Filter(pathPrefix string, filter CtxFilter) NsSection {
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	prefix := ns.name + pathPrefix
	if !ns.server.routing.MatchCase {
		prefix = strings.ToLower(prefix)
	}
	ns.setFilter(prefix, filter)
	return ns
}

func (ns *namespace) GetSetting(key string) string {
	v, ok := ns.settings[key]
	if ok {
		return v
	}
	return ""
}

func (ns *namespace) StaticDir(pathPrefix string) NsSection {
	if len(pathPrefix) < 1 {
		panic(errPathPrefix)
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	ns.server.staticDir(ns.Name() + pathPrefix)
	return ns
}

func (ns *namespace) StaticFile(file string) NsSection {
	if len(file) < 1 {
		panic(errPathPrefix)
	}
	if !strings.HasPrefix(file, "/") {
		file = "/" + file
	}
	if strings.HasSuffix(file, "/") {
		panic(errInvalidPath)
	}
	ns.server.staticFile(ns.Name() + file)
	return ns
}

func (ns *namespace) AddViewFunc(name string, f interface{}) NsSection {
	ns.addViewFunc(name, f)
	return ns
}

func (ns *namespace) RenderView(viewName string, data interface{}) ([]byte, error) {
	return ns.renderView(viewName, data)
}

func (ns *namespace) nsSettingFile() string {
	return ns.server.mapPath(ns.Name() + "/settings.xml")
}

func (ns *namespace) isConfigFile(f string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ns.nsSettingFile(), f)
	} else {
		return ns.nsSettingFile() == f
	}
}

func (ns *namespace) isInViewFolder(f string) bool {
	var viewPath = ns.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

func (ns *namespace) loadConfig() {
	var path = ns.nsSettingFile()
	if IsFile(path) {
		var settings = &nsSettingGroup{}
		//ns.server.logWriter().Println("load config file '" + path + "' for namespace '" + ns.GetName() + "'")
		err := file2Xml(path, settings)
		if err != nil {
			return
		}
		settingMap := make(map[string]string)
		for _, s := range settings.Settings {
			settingMap[s.Key] = s.Value
		}
		ns.settings = settingMap
	}
}

func (ns *namespace) viewFolder() string {
	return ns.server.mapPath(ns.Name() + "/views")
}
