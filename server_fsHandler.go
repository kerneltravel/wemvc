package wemvc

import (
	"fsnotify"
	"path"
	"strings"
)

type fsConfigHandler struct {
	app *server
}

func (d *fsConfigHandler) CanHandle(path string) bool {
	return d.app.isConfigFile(path)
}

func (d *fsConfigHandler) Handle(ev *fsnotify.Event) {
	strFile := path.Clean(ev.Name)
	conf, err := newConfig(strFile)
	if err == nil {
		d.app.config = conf
		d.app.internalErr = nil
	} else {
		d.app.internalErr = err
	}
}

type fsNsConfigHandler struct {
	app *server
	ns  *NsSection
}

func (d *fsNsConfigHandler) CanHandle(path string) bool {
	for _, ns := range app.namespaces {
		if ns.isConfigFile(path) {
			d.ns = ns
			return true
		}
	}
	return false
}

func (d *fsNsConfigHandler) Handle(ev *fsnotify.Event) {
	d.ns.loadConfig()
}

type fsViewHandler struct {
	app *server
}

func (d *fsViewHandler) CanHandle(path string) bool {
	return d.app.isInViewFolder(path)
}

func (d *fsViewHandler) Handle(ev *fsnotify.Event) {
	strFile := path.Clean(ev.Name)
	lowerStrFile := strings.ToLower(strFile)
	if IsDir(strFile) {
		if ev.Op&fsnotify.Remove == fsnotify.Remove {
			d.app.fileWatcher.RemoveWatch(strFile)
		} else if ev.Op&fsnotify.Create == fsnotify.Create {
			d.app.fileWatcher.AddWatch(strFile)
		}
	} else if strings.HasSuffix(lowerStrFile, ".html") {
		d.app.compileViews(d.app.viewFolder())
	}
}

type fsNsViewHandler struct {
	app *server
	ns  *NsSection
}

func (d *fsNsViewHandler) CanHandle(path string) bool {
	for _, ns := range d.app.namespaces {
		if ns.isInViewFolder(path) {
			d.ns = ns
			return true
		}
	}
	return false
}

func (d *fsNsViewHandler) Handle(ev *fsnotify.Event) {
	strFile := path.Clean(ev.Name)
	lowerStrFile := strings.ToLower(strFile)
	if IsDir(strFile) {
		if ev.Op&fsnotify.Remove == fsnotify.Remove {
			d.app.fileWatcher.RemoveWatch(strFile)
		} else if ev.Op&fsnotify.Create == fsnotify.Create {
			d.app.fileWatcher.AddWatch(strFile)
		}
	} else if strings.HasSuffix(lowerStrFile, ".html") {
		d.ns.compileViews(d.ns.viewFolder())
	}
}
