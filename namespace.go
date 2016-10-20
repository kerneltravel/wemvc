package wemvc

import (
	"strings"
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
	views    map[string]*view
	server   *server
	settings map[string]string
	viewContainer
	filterContainer
}

func (ns *namespace) GetName() string {
	return ns.name
}

func (ns *namespace) GetNsDir() string {
	return ns.server.MapPath(ns.GetName())
}

func (ns *namespace) Route(routePath string, c interface{}, defaultAction ...string) NsSection {
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
	ns.server.StaticDir(ns.GetName() + pathPrefix)
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
	ns.server.StaticFile(ns.GetName() + file)
	return ns
}

func (ns *namespace) AddViewFunc(name string, f interface{}) NsSection {
	ns.addViewFunc(name, f)
	//ns.server.logWriter().Println("add view func:", name)
	return ns
}

func (ns *namespace) nsSettingFile() string {
	return ns.server.MapPath(ns.GetName() + "/settings.xml")
}

func (ns *namespace) isConfigFile(f string) bool {
	return ns.nsSettingFile() == f
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
	return ns.server.MapPath(ns.GetName() + "/views")
}
