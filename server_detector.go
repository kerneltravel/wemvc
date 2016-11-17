package wemvc

import (
	"github.com/howeyc/fsnotify"
	"path"
	"strings"
)

type configDetector struct {
	app *server
}

func (d *configDetector) CanHandle(path string) bool {
	return d.app.isConfigFile(path)
}

func (d *configDetector) Handle(ev *fsnotify.FileEvent) {
	strFile := path.Clean(ev.Name)
	conf, err := newConfig(strFile)
	if err == nil {
		d.app.config = conf
		d.app.internalErr = nil
	} else {
		d.app.internalErr = err
	}
}

type nsConfigDetector struct {
	app *server
	ns *NsSection
}

func (d *nsConfigDetector) CanHandle(path string) bool {
	for _, ns := range app.namespaces {
		if ns.isConfigFile(path) {
			d.ns = ns
			return true
		}
	}
	return false
}

func (d *nsConfigDetector) Handle(ev *fsnotify.FileEvent) {
	d.ns.loadConfig()
}

type viewDetector struct {
	app *server
}

func (d *viewDetector) CanHandle(path string) bool {
	return d.app.isInViewFolder(path)
}

func (d *viewDetector) Handle(ev *fsnotify.FileEvent) {
	strFile := path.Clean(ev.Name)
	lowerStrFile := strings.ToLower(strFile)
	if IsDir(strFile) {
		if ev.IsDelete() {
			d.app.fileWatcher.RemoveWatch(strFile)
		} else if ev.IsCreate() {
			d.app.fileWatcher.AddWatch(strFile)
		}
	} else if strings.HasSuffix(lowerStrFile, ".html") {
		d.app.compileViews(d.app.viewFolder())
	}
}

type nsViewDetector struct {
	app *server
	ns *NsSection
}

func (d *nsViewDetector) CanHandle(path string) bool {
	for _, ns := range d.app.namespaces {
		if ns.isInViewFolder(path) {
			d.ns = ns
			return true
		}
	}
	return false
}

func (d *nsViewDetector) Handle(ev *fsnotify.FileEvent) {
	strFile := path.Clean(ev.Name)
	lowerStrFile := strings.ToLower(strFile)
	if IsDir(strFile) {
		if ev.IsDelete() {
			d.app.fileWatcher.RemoveWatch(strFile)
		} else if ev.IsCreate() {
			d.app.fileWatcher.AddWatch(strFile)
		}
	} else if strings.HasSuffix(lowerStrFile, ".html") {
		d.ns.compileViews(d.ns.viewFolder())
	}
}