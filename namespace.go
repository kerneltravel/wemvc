package wemvc

import (
	"strings"
	"runtime"
)

type nsSettingGroup struct {
	Settings []struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	} `xml:"add"`
}

type NsSection struct {
	name     string
	server   *server
	settings map[string]string
	viewContainer
	filterContainer
}

func (ns *NsSection) Name() string {
	return ns.name
}

func (ns *NsSection) Dir() string {
	return ns.server.mapPath(ns.Name())
}

func (ns *NsSection) Route(routePath string, c interface{}, defaultAction ...string) *NsSection {
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

func (ns *NsSection) Filter(pathPrefix string, filter CtxFilter) *NsSection {
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

func (ns *NsSection) GetSetting(key string) string {
	v, ok := ns.settings[key]
	if ok {
		return v
	}
	return ""
}

func (ns *NsSection) StaticDir(pathPrefix string) *NsSection {
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

func (ns *NsSection) StaticFile(file string) *NsSection {
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

func (ns *NsSection) AddViewFunc(name string, f interface{}) *NsSection {
	ns.addViewFunc(name, f)
	return ns
}

func (ns *NsSection) RenderView(viewName string, data interface{}) ([]byte, error) {
	return ns.renderView(viewName, data)
}

func (ns *NsSection) nsSettingFile() string {
	return ns.server.mapPath(ns.Name() + "/settings.xml")
}

func (ns *NsSection) isConfigFile(f string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(ns.nsSettingFile(), f)
	} else {
		return ns.nsSettingFile() == f
	}
}

func (ns *NsSection) isInViewFolder(f string) bool {
	var viewPath = ns.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

func (ns *NsSection) loadConfig() {
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

func (ns *NsSection) viewFolder() string {
	return ns.server.mapPath(ns.Name() + "/views")
}
