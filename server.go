package wemvc

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/howeyc/fsnotify"

	"container/list"
	"net/url"
	"runtime"
)

type server struct {
	errorHandlers   map[int]CtxHandler
	port            int
	webRoot         string
	config          *config
	routing         *routeTree
	locked          bool
	staticPaths     []string
	staticFiles     []string
	globalSession   *SessionManager
	namespaces      map[string]*namespace
	sessionProvides map[string]SessionProvider
	globalFilters   []CtxFilter
	internalErr     error
	fileWatcher     *FileWatcher
	viewContainer
	filterContainer
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
func (app *server) mapPath(virtualPath string) string {
	var res = path.Join(app.webRoot, virtualPath)
	return fixPath(res)
}

// ServeHTTP serve the
func (app *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// handle 500 errors
	defer app.panicRecover(w, req)
	if app.internalErr != nil {
		panic(app.internalErr)
	}
	// Init context obj
	var ctx = &Context{
		req: req,
		w:   w,
		app: app,
	}
	// execute global filters
	if len(app.globalFilters) > 0 {
		for _, filter := range app.globalFilters {
			filter(ctx)
			if ctx.ended {
				break
			}
		}
	}
	// flush the request
	app.flushRequest(w, req, ctx.Result)
}

func (app *server) staticDir(pathPrefix string) {
	app.assertNotLocked()
	if len(pathPrefix) < 1 {
		panic(errPathPrefix)
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	if runtime.GOOS == "windows" {
		pathPrefix = strings.ToLower(pathPrefix)
	}
	app.staticPaths = append(app.staticPaths, pathPrefix)
}

func (app *server) staticFile(path string) {
	app.assertNotLocked()
	if len(path) < 1 {
		panic(errPathPrefix)
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(path, "/") {
		panic(errInvalidPath)
	}
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	app.staticFiles = append(app.staticFiles, path)
}

func (app *server) getNamespace(nsName string) NsSection {
	if len(nsName) > 0 {
		if !strings.HasPrefix(nsName, "/") {
			nsName = "/" + nsName
		}
		nsName = strings.TrimRight(nsName, "/")
	}
	if len(nsName) < 1 {
		panic(errInvalidNamespace)
	}
	if app.namespaces == nil {
		app.namespaces = make(map[string]*namespace)
	}
	ns, ok := app.namespaces[nsName]
	if ok {
		return ns
	}
	ns = &namespace{
		name:   nsName,
		server: app,
	}
	ns.viewExt = app.viewExt
	app.namespaces[nsName] = ns
	return ns
}

func (app *server) assertNotLocked() {
	if app.locked {
		panic("Invalid operation. You cannot call this function after the server setting is locked")
	}
}

func (app *server) addRoute(namespace string, routePath string, c interface{}, action string) {
	t := reflect.TypeOf(c)
	cInfo := newControllerInfo(app, namespace, t, action)
	if app.routing == nil {
		app.routing = newRouteTree()
	}
	app.routing.addRoute(routePath, cInfo)
}

func (app *server) flushRequest(w http.ResponseWriter, req *http.Request, result interface{}) {
	if result == nil {
		result = app.handleErrorReq(req, 404)
	}
	switch result.(type) {
	case *FileResult:
		http.ServeFile(w, req, result.(*FileResult).FilePath)
		return
	case *RedirectResult:
		res := result.(*RedirectResult)
		if res.StatusCode != 301 {
			res.StatusCode = 302
		}
		http.Redirect(w, req, res.RedirectUrl, res.StatusCode)
		return
	case *url.URL:
		res := result.(*url.URL)
		http.Redirect(w, req, res.String(), 302)
		return
	case *Result:
		res := result.(*Result)
		// write the result to browser
		for k, v := range res.Headers {
			if k == "Content-Type" {
				continue
			}
			w.Header().Add(k, v)
		}
		contentType := fmt.Sprintf("%s;charset=%s", res.ContentType, res.Encoding)
		w.Header().Add("Content-Type", contentType)
		if res.StatusCode != 200 {
			w.WriteHeader(res.StatusCode)
		}
		output := res.GetOutput()
		if len(output) > 0 {
			w.Write(res.GetOutput())
		}
		return
	case string:
	case []byte:
		w.Header().Add("Content-Type", "text/plain")
		w.Write(result.([]byte))
		return
	case byte:
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte{result.(byte)})
		return
	default:
		var cType = req.Header.Get("Content-Type")
		var contentBytes []byte
		var err error
		if cType == "text/xml" {
			contentBytes, err = xml.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "text/xml")
		} else {
			contentBytes, err = json.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "application/json")
		}
		w.Write(contentBytes)
	}
}

// init used to initialize the server
// 1. load the config file
// 2. watch the view file
// 3. init the global session
func (app *server) init() error {
	// init the error handler
	app.errorHandlers[404] = app.error404
	app.errorHandlers[403] = app.error403
	// fsnotify watcher
	w, err := NewWatcher()
	if err != nil {
		return err
	}
	app.fileWatcher = w
	app.addWatcherHandler()
	app.fileWatcher.Start()

	// load & watch the global config files
	globalConfigFile := app.mapPath("/config.xml")
	conf,err := newConfig(globalConfigFile)
	app.config = conf
	err1 := app.fileWatcher.AddWatch(globalConfigFile)
	if err != nil {
		return err
	}
	if err1 != nil {
		panic(err1)
	}
	// build the view template and watch the changes
	viewDir := app.viewFolder()
	if IsDir(viewDir) {
		app.compileViews(viewDir)
		//app.logWriter().Println("compile view files in dir", viewDir)
		app.fileWatcher.AddWatch(viewDir)
		filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error {
			if info.IsDir() {
				app.fileWatcher.AddWatch(p)
			}
			return nil
		})
	}
	// process namespaces: build the views files and load the config
	if app.namespaces != nil {
		//for name, ns := range app.namespaces {
		for _, ns := range app.namespaces {
			//app.logWriter().Println("process namespace", name)
			settingFile := ns.nsSettingFile()
			ns.loadConfig()
			app.fileWatcher.AddWatch(settingFile)
			nsViewDir := ns.viewFolder()
			ns.compileViews(nsViewDir)
			//app.logWriter().Println("compile view files in dir", nsViewDir)
			app.fileWatcher.AddWatch(nsViewDir)
			filepath.Walk(nsViewDir, func(p string, info os.FileInfo, er error) error {
				if er != nil {
					return nil
				}
				if info != nil && info.IsDir() {
					app.fileWatcher.AddWatch(p)
				}
				return nil
			})
		}
	}

	// init sessionManager
	mgr, err := app.NewSessionManager(app.config.SessionConfig.ManagerName, app.config.SessionConfig)
	if err != nil {
		return err
	}
	app.globalSession = mgr
	go app.globalSession.GC()
	return nil
}

// watchFile used to watching the required files: config files and view files
func (app *server) addWatcherHandler() {
	if app.fileWatcher == nil {
		return
	}
	// add config file handler
	app.fileWatcher.AddHandler(&configDetector{app:app}, func(ev *fsnotify.FileEvent) bool {
		strFile := path.Clean(ev.Name)
		conf,err := newConfig(strFile)
		if err == nil {
			app.config = conf
			app.internalErr = nil
		} else {
			app.internalErr = err
		}
		return false;
	})
	// add ns config handler
	app.fileWatcher.AddHandler(&nsConfigDetector{app:app}, func(ev *fsnotify.FileEvent) bool {
		for _, ns := range app.namespaces {
			strFile := path.Clean(ev.Name)
			if ns.isConfigFile(strFile) {
				ns.loadConfig()
			}
		}
		return false
	})
	// add view file handler
	app.fileWatcher.AddHandler(&viewDetector{app:app}, func(ev *fsnotify.FileEvent) bool {
		strFile := path.Clean(ev.Name)
		lowerStrFile := strings.ToLower(strFile)
		if IsDir(strFile) {
			if ev.IsDelete() {
				app.fileWatcher.RemoveWatch(strFile)
			} else if ev.IsCreate() {
				app.fileWatcher.AddWatch(strFile)
			}
		} else if strings.HasSuffix(lowerStrFile, ".html") {
			app.compileViews(app.viewFolder())
		}
		return false
	})
	// add ns view file handler
	app.fileWatcher.AddHandler(&nsViewDetector{app:app}, func(ev *fsnotify.FileEvent) bool {
		strFile := path.Clean(ev.Name)
		lowerStrFile := strings.ToLower(strFile)
		for _, ns := range app.namespaces {
			if ns.isInViewFolder(strFile) {
				if IsDir(strFile) {
					if ev.IsDelete() {
						app.fileWatcher.RemoveWatch(strFile)
					} else if ev.IsCreate() {
						app.fileWatcher.AddWatch(strFile)
					}
				} else if strings.HasSuffix(lowerStrFile, ".html") {
					ns.compileViews(ns.viewFolder())
				}
				return false
			}
		}
		return false
	})
}

func (app *server) isConfigFile(f string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(app.mapPath("/config.xml"), f)
	} else {
		return app.mapPath("/config.xml") == f
	}
}

func (app *server) isNsConfigFile(f string) bool {
	for _, ns := range app.namespaces {
		if ns.isConfigFile(f) {
			return true
		}
	}
	return false
}

func (app *server) isInViewFolder(f string) bool {
	viewPath := app.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

// isStaticRequest check the current request is indicate to static path
func (app *server) isStaticRequest(req *http.Request) bool {
	var url string
	if runtime.GOOS == "windows" {
		url = strings.ToLower(req.URL.Path)
	} else {
		url = req.URL.Path
	}
	for _, f := range app.staticFiles {
		if f == url {
			return true
		}
	}
	for _, p := range app.staticPaths {
		if strings.HasPrefix(url, p) {
			return true
		}
	}
	return false
}

func (app *server) viewFolder() string {
	return app.mapPath("/views")
}

func newServer(webRoot string) *server {
	var app = &server{
		webRoot:       webRoot,
		locked:        false,
		errorHandlers: make(map[int]CtxHandler),
	}
	app.views = make(map[string]*view)
	app.filters = make(map[string][]CtxFilter)
	app.viewExt = ".html"
	app.sessionProvides = make(map[string]SessionProvider)
	app.RegSessionProvider("memory", &MemSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)})
	app.globalFilters = []CtxFilter{
		ServeStatic,
		InitRoute,
		HandleRoute,
		ExecutePathFilters,
		ExecuteAction,
	}
	return app
}
