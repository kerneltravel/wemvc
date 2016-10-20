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
	"runtime/debug"

	"container/list"
	"runtime"
)

// Server the application interface that define the useful function
type Server interface {
	RootDir() string
	Config() Configuration
	MapPath(virtualPath string) string
	//Logger() *log.Logger
	Namespace(ns string) NsSection
	SetRootDir(rootDir string)
	StaticDir(pathPrefix string)
	StaticFile(path string)
	HandleError(errorCode int, handler CtxHandler)
	Route(routePath string, c interface{}, defaultAction ...string)
	SetPathFilter(pathPrefix string, filter CtxFilter)
	SetGlobalFilter(filters []CtxFilter)
	//SetLogFile(name string) Server
	SetViewExt(ext string)
	AddViewFunc(name string, f interface{})
	AddRouteFunc(name string, fun RouteValidateFunc)
	RegSessionProvider(name string, provide SessionProvider)
	NewSessionManager(provideName string, config *SessionConfig) (*SessionManager, error)
	Run(port int)
	RunTLS(port int, certFile, keyFile string)
}

type server struct {
	errorHandlers map[int]CtxHandler
	port          int
	webRoot       string
	config        *config
	routing       *routeTree
	watcher       *fsnotify.Watcher
	locked        bool
	staticPaths   []string
	staticFiles   []string
	globalSession *SessionManager
	//logger          *log.Logger
	namespaces      map[string]*namespace
	sessionProvides map[string]SessionProvider
	globalFilters   []CtxFilter
	viewContainer
	filterContainer
}

// RootDir get the root file path of the web server
func (app *server) RootDir() string {
	return app.webRoot
}

// Config get the config data
func (app *server) Config() Configuration {
	return app.config
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
func (app *server) MapPath(virtualPath string) string {
	var res = path.Join(app.RootDir(), virtualPath)
	return fixPath(res)
}

// SetRootDir set the root directory of the web application
func (app *server) SetRootDir(rootDir string) {
	app.assertNotLocked()
	if !IsDir(rootDir) {
		panic(errInvalidRoot)
	}
	app.webRoot = rootDir
	//return app
}

// SetViewExt set the view file extension
func (app *server) SetViewExt(ext string) {
	app.assertNotLocked()
	if len(ext) < 1 || !strings.HasPrefix(ext, ".") {
		return
	}
	if runtime.GOOS == "windows" {
		app.viewExt = strings.ToLower(ext)
	} else {
		app.viewExt = ext
	}
	if app.namespaces != nil {
		for _, ns := range app.namespaces {
			ns.viewDir = app.viewDir
		}
	}
	//return app
}

// ServeHTTP serve the
func (app *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// handle 500 errors
	defer app.panicRecover(w, req)
	var ctx = &Context{
		req: req,
		w:   w,
		app: app,
	}
	if len(app.globalFilters) > 0 {
		for _, filter := range app.globalFilters {
			filter(ctx)
			if ctx.ended {
				break
			}
		}
	}
	// process the dynamic result
	app.flushRequest(w, req, ctx.Result)
}

// AddStatic set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func (app *server) StaticDir(pathPrefix string) {
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

func (app *server) StaticFile(path string) {
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

func (app *server) HandleError(errorCode int, handler CtxHandler) {
	app.assertNotLocked()
	app.errorHandlers[errorCode] = handler
}

func (app *server) AddViewFunc(name string, f interface{}) {
	app.assertNotLocked()
	app.addViewFunc(name, f)
}

func (app *server) AddRouteFunc(name string, f RouteValidateFunc) {
	app.assertNotLocked()
	err := app.routing.addFunc(name, f)
	if err != nil {
		panic(err)
	}
}

func (app *server) Route(routePath string, c interface{}, defaultAction ...string) {
	app.assertNotLocked()
	action := "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	app.route("", routePath, c, action)
}

func (app *server) SetPathFilter(pathPrefix string, filter CtxFilter) {
	app.assertNotLocked()
	if !app.routing.MatchCase {
		pathPrefix = strings.ToLower(pathPrefix)
	}
	app.setFilter(pathPrefix, filter)
}

func (app *server) SetGlobalFilter(filters []CtxFilter) {
	app.assertNotLocked()
	if len(filters) < 1 {
		return
	}
	app.globalFilters = filters
}

func (app *server) Namespace(nsName string) NsSection {
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

func (app *server) Run(port int) {
	//app.logWriter().Println("use root dir '" + app.webRoot + "'")
	err := app.init()
	if err != nil {
		//app.logWriter().Println(err.Error())
		return
	}
	app.locked = true
	app.port = port
	//host, err := os.Hostname()
	//if err != nil {
	//	host = "localhost"
	//}
	//app.logWriter().Println(fmt.Sprintf("server is running on port '%d'. http://%s:%d", app.port, host, app.port))
	portStr := fmt.Sprintf(":%d", app.port)
	err = http.ListenAndServe(portStr, app)
	if err != nil {
		//app.logWriter().Println(err.Error())
		panic(err)
	}
}

func (app *server) RunTLS(port int, certFile, keyFile string) {
	//app.logWriter().Println("use root dir '" + app.webRoot + "'")
	err := app.init()
	if err != nil {
		//app.logWriter().Println(err.Error())
		return
	}
	app.locked = true
	app.port = port
	//host, err := os.Hostname()
	//if err != nil {
	//	host = "localhost"
	//}
	//app.logWriter().Println(fmt.Sprintf("server is running on port '%d'. http://%s:%d", app.port, host, app.port))
	portStr := fmt.Sprintf(":%d", app.port)
	err = http.ListenAndServeTLS(portStr, certFile, keyFile, app)
	if err != nil {
		//app.logWriter().Println(err.Error())
		panic(err)
	}
}

func (app *server) assertNotLocked() {
	if app.locked {
		panic("Invalid operation. You cannot call this function after the server setting is locked")
	}
}

func (app *server) route(namespace string, routePath string, c interface{}, action string) {
	t := reflect.TypeOf(c)
	cInfo := newControllerInfo(app, namespace, t, action)
	if app.routing == nil {
		app.routing = newRouteTree()
	}
	//app.logWriter().Println("set route '"+routePath+"'        controller:", cInfo.CtrlType.Name(), "       default action:", cInfo.DefaultAction+"\r\n")
	app.routing.addRoute(routePath, cInfo)
}

func (app *server) flushRequest(w http.ResponseWriter, req *http.Request, result interface{}) {
	if result == nil {
		result = app.handleError(req, 404)
	}
	switch result.(type) {
	case FileResult:
		http.ServeFile(w, req, result.(FileResult).FilePath)
		return
	case *FileResult:
		http.ServeFile(w, req, result.(*FileResult).FilePath)
		return
	case RedirectResult:
		res := result.(RedirectResult)
		if res.StatusCode != 301 {
			res.StatusCode = 302
		}
		http.Redirect(w, req, res.RedirectUrl, res.StatusCode)
		return
	case *RedirectResult:
		res := result.(*RedirectResult)
		if res.StatusCode != 301 {
			res.StatusCode = 302
		}
		http.Redirect(w, req, res.RedirectUrl, res.StatusCode)
		return
	case Result:
		res := result.(Result)
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
		if cType == "text/xml" {
			xmlBytes, err := xml.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "text/xml")
			w.Write(xmlBytes)
		} else {
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				panic(err)
			}
			w.Header().Add("Content-Type", "application/json")
			w.Write(jsonBytes)
		}
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
	// init fsnotify watcher
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	app.watcher = w
	// load & watch the global config files
	globalConfigFile := app.MapPath("/config.xml")
	config := &config{svr: app}
	if config.loadFile(globalConfigFile) {
		err = app.watcher.Watch(globalConfigFile)
		if err != nil {
			panic(err)
		}
	}
	app.config = config
	// build the view template and watch the changes
	viewDir := app.viewFolder()
	if IsDir(viewDir) {
		app.compileViews(viewDir)
		//app.logWriter().Println("compile view files in dir", viewDir)
		app.watcher.Watch(viewDir)
		filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error {
			if info.IsDir() {
				app.watcher.Watch(p)
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
			app.watcher.Watch(settingFile)
			nsViewDir := ns.viewFolder()
			ns.compileViews(nsViewDir)
			//app.logWriter().Println("compile view files in dir", nsViewDir)
			app.watcher.Watch(nsViewDir)
			filepath.Walk(nsViewDir, func(p string, info os.FileInfo, er error) error {
				if er != nil {
					return nil
				}
				if info != nil && info.IsDir() {
					app.watcher.Watch(p)
				}
				return nil
			})
		}
	}
	// start to watch the files and dirs
	go app.watchFile()
	// init sessionManager
	mgr, err := app.NewSessionManager(app.config.SessionConfig.ManagerName, app.config.SessionConfig)
	if err != nil {
		panic(err)
	}
	app.globalSession = mgr
	go app.globalSession.GC()
	return nil
}

// watchFile used to watching the required files: config files and view files
func (app *server) watchFile() {
	for {
		select {
		case ev := <-app.watcher.Event:
			strFile := path.Clean(ev.Name)
			lowerStrFile := strings.ToLower(strFile)
			if app.isConfigFile(strFile) {
				//app.logWriter().Println("config file", strFile, "has been changed")
				conf := &config{svr: app}
				if conf.loadFile(strFile) {
					app.config = conf
				}
			} else if app.isNsConfigFile(strFile) {
				for _, ns := range app.namespaces {
					if ns.isConfigFile(strFile) {
						ns.loadConfig()
					}
				}
			} else {
				//app.logWriter().Println("view file", strFile, "has been changed")
				for _, ns := range app.namespaces {
					if ns.isInViewFolder(strFile) {
						if IsDir(strFile) {
							if ev.IsDelete() {
								app.watcher.RemoveWatch(strFile)
							} else if ev.IsCreate() {
								app.watcher.Watch(strFile)
							}
						} else if strings.HasSuffix(lowerStrFile, ".html") {
							ns.compileViews(ns.viewFolder())
							//app.logWriter().Println("compile view files in dir", ns.viewFolder())
						}
						break
					}
				}
				if app.isInViewFolder(strFile) {
					if IsDir(strFile) {
						if ev.IsDelete() {
							app.watcher.RemoveWatch(strFile)
						} else if ev.IsCreate() {
							app.watcher.Watch(strFile)
						}
					} else if strings.HasSuffix(lowerStrFile, ".html") {
						app.compileViews(app.viewFolder())
						//app.logWriter().Println("compile view files in dir", app.viewFolder())
					}
				}
			}
		}
	}
}

func (app *server) isConfigFile(f string) bool {
	if app.MapPath("/config.xml") == f {
		return true
	}
	return false
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
func (app *server) isStaticRequest(url string) bool {
	if runtime.GOOS == "windows" {
		url = strings.ToLower(url)
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

func (app *server) error404(req *http.Request) *Result {
	return renderError(
		404,
		"The resource you are looking for has been removed, had its name changed, or is temporarily unavailable",
		"Request URL: "+req.URL.String(),
		"",
	)
}

func (app *server) error403(req *http.Request) *Result {
	return renderError(
		403,
		"The server understood the request but refuses to authorize it",
		"Request URL: "+req.URL.String(),
		"",
	)
}

func (app *server) handleError(req *http.Request, code int, title ...string) *Result {
	var handler = app.errorHandlers[code]
	if handler != nil {
		return handler(req)
	} else if errTitle, ok := statusCodeMapping[code]; ok {
		t := errTitle
		if len(title) > 0 && len(title[0]) > 0 {
			t = t + ":" + title[0]
		}
		return renderError(
			code,
			t,
			"Request URL: "+req.URL.String(),
			"",
		)
	}
	return app.error404(req)
}

func (app *server) viewFolder() string {
	return app.MapPath("/views")
}

func (app *server) panicRecover(res http.ResponseWriter, req *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	// detect end request
	_, ok := rec.(*errEndRequest)
	if ok {
		return
	}
	// process 500 error
	res.WriteHeader(500)
	var debugStack = string(debug.Stack())
	debugStack = strings.Replace(debugStack, "<", "&lt;", -1)
	debugStack = strings.Replace(debugStack, ">", "&gt;", -1)
	if err, ok := rec.(error); ok {
		res.Write(genError(500, "", err.Error(), debugStack))
	} else {
		res.Write(genError(500, "", "Unkown Internal Server Error", debugStack))
	}
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
