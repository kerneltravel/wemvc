package wemvc

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"runtime/debug"

	"errors"

	"github.com/Simbory/wemvc/fsnotify"
	"github.com/Simbory/wemvc/utils"
	"runtime"
)

// Server the application interface that define the useful function
type Server interface {
	RootDir() string
	Config() Configuration
	MapPath(virtualPath string) string
	Logger() *log.Logger
	Namespace(ns string) NamespaceSection
	SetRootDir(rootDir string) Server
	StaticDir(pathPrefix string) Server
	StaticFile(path string) Server
	HandleError(errorCode int, handler CtxHandler) Server
	Route(routePath string, c interface{}, defaultAction ...string) Server
	Filter(pathPrefix string, filter FilterFunc) Server
	SetLogFile(name string) Server
	SetViewExt(ext string) Server
	AddViewFunc(name string, f interface{}) Server
	Run(port int) error
}

type server struct {
	errorHandlers map[int]CtxHandler
	port          int
	webRoot       string
	config        *config
	router        *router
	watcher       *fsnotify.Watcher
	routeLocked   bool
	staticPaths   []string
	staticFiles   []string
	globalSession *SessionManager
	logger        *log.Logger
	namespaces    map[string]*namespace
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

// SetRootDir set the root directory of the web application
func (app *server) SetRootDir(rootDir string) Server {
	if app.routeLocked {
		panic(errors.New("Cannot set the web root while the application is running."))
	}
	if !utils.IsDir(rootDir) {
		panic("invalid root dir")
	}
	app.webRoot = rootDir
	return app
}

// SetViewExt set the view file extension
func (app *server) SetViewExt(ext string) Server {
	if app.routeLocked {
		return app
	}
	if len(ext) < 1 || !strings.HasPrefix(ext, ".") {
		return app
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
	return app
}

// ServeHTTP serve the
func (app *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// handle 500 errors
	defer app.panicRecover(w, req)

	// var lowerURL = strings.ToLower(req.URL.Path)
	var ctx = &context{
		req: req,
		w:   w,
		end: false,
	}
	var result ActionResult
	// serve the static file
	if app.isStaticRequest(req.URL.Path) {
		app.serveStaticFile(ctx)
		if ctx.end {
			ctx = nil
			return
		}
	} else {
		// serve the dynamic page
		ctx = app.execRoute(ctx)
		if ctx != nil {
			// execute the global filters
			if len(ctx.ns) < 1 {
				if app.execFilters(ctx) {
					return
				}
			} else {
				ns, ok := app.namespaces[ctx.ns]
				if ok && ns != nil {
					if ns.execFilters(ctx) {
						return
					}
				} else {
					result = nil
					goto error404
				}
			}
			result = app.handleDynamic(ctx)
		}
	}
	// handle error 404
error404:
	if result == nil {
		result = app.handleError(req, 404)
	}
	// process the dynamic result
	app.flushRequest(result, w, req)
	result = nil
}

// AddStatic set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func (app *server) StaticDir(pathPrefix string) Server {
	if len(pathPrefix) < 1 {
		panic(errors.New("the static path prefix cannot be empty"))
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
	return app
}

func (app *server) StaticFile(path string) Server {
	if len(path) < 1 {
		panic(errors.New("the static path prefix cannot be empty"))
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if strings.HasSuffix(path, "/") {
		panic(errors.New("the static file path cannot be end with '/'"))
	}
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	app.staticFiles = append(app.staticFiles, path)
	return app
}

func (app *server) HandleError(errorCode int, handler CtxHandler) Server {
	app.errorHandlers[errorCode] = handler
	return app
}

func (app *server) AddViewFunc(name string, f interface{}) Server {
	app.addViewFunc(name, f)
	return app
}

func (app *server) Route(routePath string, c interface{}, defaultAction ...string) Server {
	var action = "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	app.route("", routePath, c, action)
	return app
}

func (app *server) Filter(pathPrefix string, filter FilterFunc) Server {
	app.setFilter(pathPrefix, filter)
	return app
}

func (app *server) Logger() *log.Logger {
	return app.logWriter()
}

func (app *server) SetLogFile(name string) Server {
	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err.Error())
		return app
	}
	logger := log.New(file, "", log.LstdFlags|log.Llongfile)
	app.logger = logger
	return app
}

func (app *server) Namespace(nsName string) NamespaceSection {
	if len(nsName) > 0 {
		if !strings.HasPrefix(nsName, "/") {
			nsName = "/" + nsName
		}
		nsName = strings.TrimRight(nsName, "/")
	}
	if len(nsName) < 1 {
		panic("invalid namespace")
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

func (app *server) Run(port int) error {
	app.logWriter().Println("use root dir '" + app.webRoot + "'")
	err := app.init()
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
	app.routeLocked = true
	app.port = port
	host, err := os.Hostname()
	if err != nil {
		host = "localhost"
	}
	app.logWriter().Println(fmt.Sprintf("server is running on port '%d'. http://%s:%d", app.port, host, app.port))
	portStr := fmt.Sprintf(":%d", app.port)
	return http.ListenAndServe(portStr, app)
}

func (app *server) route(namespace string, routePath string, c interface{}, action string) {
	if app.routeLocked {
		panic(errors.New("This route cannot be added, because the route table is locked."))
	}
	var t = reflect.TypeOf(c)
	cInfo := newControllerInfo(namespace, t, action)
	if app.router == nil {
		app.router = newRouter()
	}
	app.logWriter().Println("set route '"+routePath+"'        controller:", cInfo.controllerType.Name(), "       default action:", cInfo.defaultAction+"\r\n")
	app.router.handle(routePath, cInfo)
}

func (app *server) flushRequest(result ActionResult, w http.ResponseWriter, req *http.Request) {
	res, ok := result.(*actionResult)
	if ok {
		if len(res.resFile) > 0 {
			http.ServeFile(w, req, res.resFile)
			return
		}
		if len(res.redURL) > 0 {
			http.Redirect(w, req, res.redURL, res.statusCode)
			return
		}
	}
	// write the result to browser
	for k, v := range result.GetHeaders() {
		w.Header().Add(k, v)
	}
	var contentType = fmt.Sprintf("%s;charset=%s", result.GetContentType(), result.GetEncoding())
	w.Header().Add("Content-Type", contentType)
	if result.GetStatusCode() != 200 {
		w.WriteHeader(result.GetStatusCode())
	}
	var output = result.GetOutput()
	if len(output) > 0 {
		w.Write(result.GetOutput())
	}
}

// mapPath Returns the physical file path that corresponds to the specified virtual path.
func (app *server) MapPath(virtualPath string) string {
	var res = path.Join(RootDir(), virtualPath)
	return utils.FixPath(res)
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
	var globalConfigFile = app.MapPath("/config.xml")
	var config = &config{svr: app}
	if config.loadFile(globalConfigFile) {
		err = app.watcher.Watch(globalConfigFile)
		if err != nil {
			panic(err)
		}
	}
	app.config = config
	// build the view template and watch the changes
	var viewDir = app.viewFolder()
	if utils.IsDir(viewDir) {
		app.compileViews(viewDir)
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
		for name, ns := range app.namespaces {
			app.logWriter().Println("process namespace", name)
			settingFile := ns.nsSettingFile()
			ns.loadConfig()
			app.watcher.Watch(settingFile)
			nsViewDir := ns.viewFolder()
			ns.compileViews(nsViewDir)
			app.watcher.Watch(nsViewDir)
			filepath.Walk(nsViewDir, func(p string, info os.FileInfo, er error) error {
				if info.IsDir() {
					app.watcher.Watch(p)
				}
				return nil
			})
		}
	}
	// start to watch the files and dirs
	go app.watchFile()
	// init sessionManager
	mgr, err := NewSessionManager(app.config.SessionConfig.ManagerName, app.config.SessionConfig)
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
				app.logWriter().Println("config file", strFile, "has been changed")
				var conf = &config{svr: app}
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
				app.logWriter().Println("view file", strFile, "has been changed")
				for _, ns := range app.namespaces {
					if ns.isInViewFolder(strFile) {
						if utils.IsDir(strFile) {
							if ev.IsDelete() {
								app.watcher.RemoveWatch(strFile)
							} else if ev.IsCreate() {
								app.watcher.Watch(strFile)
							}
						} else if strings.HasSuffix(lowerStrFile, ".html") {
							ns.compileViews(ns.viewFolder())
						}
						break
					}
				}
				if app.isInViewFolder(strFile) {
					if utils.IsDir(strFile) {
						if ev.IsDelete() {
							app.watcher.RemoveWatch(strFile)
						} else if ev.IsCreate() {
							app.watcher.Watch(strFile)
						}
					} else if strings.HasSuffix(lowerStrFile, ".html") {
						app.compileViews(app.viewFolder())
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
	var viewPath = app.viewFolder()
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

func (app *server) serveStaticFile(ctx *context) {
	var physicalFile = ""
	var f = app.MapPath(ctx.req.URL.Path)
	stat, err := os.Stat(f)
	if err != nil {
		return
	}
	if stat.IsDir() {
		absolutePath := ctx.req.URL.Path
		if !strings.HasSuffix(absolutePath, "/") {
			absolutePath = absolutePath + "/"
		}
		physicalPath := app.MapPath(absolutePath)
		if utils.IsDir(physicalPath) {
			var defaultUrls = app.config.getDefaultUrls()
			if len(defaultUrls) > 0 {
				for _, f := range defaultUrls {
					var file = app.MapPath(absolutePath + f)
					if utils.IsFile(file) {
						physicalFile = file
						break
					}
				}
			}
		}
	} else {
		physicalFile = f
	}
	if len(physicalFile) > 0 {
		app.logWriter().Println("handle static path '" + ctx.req.URL.Path + "'")
		http.ServeFile(ctx.w, ctx.req, physicalFile)
		ctx.end = true
	}
}

func (app *server) execRoute(ctx *context) *context {
	var urlPath = ctx.req.URL.Path
	if len(urlPath) > 1 && strings.HasSuffix(urlPath, "/") {
		urlPath = strings.TrimRight(urlPath, "/")
	}
	//var resp ActionResult
	cInfo, routeData, match := app.router.lookup(ctx.req.Method, urlPath)
	if !match && cInfo != nil {
		var action = routeData.ByName("action")
		var ns = cInfo.namespace
		var method = strings.ToTitle(ctx.req.Method)
		if len(action) < 1 {
			action = cInfo.defaultAction
		} else {
			action = strings.Replace(action, "-", "_", -1)
		}
		// find the action method in controller
		var actionMethod string
		if cInfo.containsAction(strings.ToLower(method + "_" + action)) {
			actionMethod = strings.ToLower(method + "_" + action)
		} else if cInfo.containsAction(strings.ToLower(method + action)) {
			actionMethod = strings.ToLower(method + action)
		} else if cInfo.containsAction(strings.ToLower(action)) {
			actionMethod = strings.ToLower(action)
		}
		if len(actionMethod) > 0 {
			actionMethod = cInfo.actions[actionMethod]
			ctx.routeData = routeData
			ctx.actionName = action
			ctx.ctrlType = cInfo.controllerType
			ctx.ns = ns
			ctx.actionMethod = actionMethod
			ctx.actionName = action
			ctx.routeData = routeData
			cName := strings.ToLower(ctx.ctrlType.String())
			cName = strings.Split(cName, ".")[1]
			cName = strings.Replace(cName, "controller", "", -1)
			ctx.ctrlName = cName
			return ctx
		}
	}
	return nil
}

func (app *server) handleDynamic(ctx *context) ActionResult {
	var ctrl = reflect.New(ctx.ctrlType)
	cAction := ctx.actionName

	// call OnInit method
	onInitMethod := ctrl.MethodByName("OnInit")
	if onInitMethod.IsValid() {
		onInitMethod.Call([]reflect.Value{
			reflect.ValueOf(ctx.req),
			reflect.ValueOf(ctx.w),
			reflect.ValueOf(ctx.ns),
			reflect.ValueOf(ctx.ctrlName),
			reflect.ValueOf(cAction),
			reflect.ValueOf(ctx.routeData),
			reflect.ValueOf(ctx.items),
		})
	}
	//parse form
	if ctx.req.Method == "POST" || ctx.req.Method == "PUT" || ctx.req.Method == "PATCH" {
		if ctx.req.MultipartForm != nil {
			var size int64
			var maxSize = app.config.GetSetting("MaxFormSize")
			if len(maxSize) < 1 {
				size = 10485760
			} else {
				size, _ = strconv.ParseInt(maxSize, 10, 64)
			}
			ctx.req.ParseMultipartForm(size)
		} else {
			ctx.req.ParseForm()
		}
	}
	// call OnLoad method
	onLoadMethod := ctrl.MethodByName("OnLoad")
	if onLoadMethod.IsValid() {
		onLoadMethod.Call(nil)
	}
	// call action method
	m := ctrl.MethodByName(ctx.actionMethod)
	if !m.IsValid() {
		return nil
	}
	if len(ctx.ns) < 1 {
		app.logWriter().Println("handle dynamic path '"+ctx.req.URL.Path+"' {\"controller\":\"", ctx.ctrlName+"\",\"action\":\""+ctx.actionName+"\"}")
	} else {
		app.logWriter().Println("handle dynamic path '"+ctx.req.URL.Path+"' {\"controller\":\"", ctx.ctrlName+"\",\"action\":\""+ctx.actionName+"\",\"namespace\":\""+ctx.ns+"\"}")
	}
	values := m.Call(nil)
	if len(values) == 1 {
		var result = values[0].Interface()
		value, valid := result.(ActionResult)
		if !valid {
			value = NewActionResult()
			var cType = ctx.req.Header.Get("Content-Type")
			if cType == "text/xml" {
				xmlBytes, err := xml.Marshal(result)
				if err != nil {
					panic(err)
				}
				value.SetContentType("text/xml")
				value.Write(xmlBytes)
			} else {
				jsonBytes, err := json.Marshal(result)
				if err != nil {
					panic(err)
				}
				value.SetContentType("application/json")
				value.Write(jsonBytes)
			}
		}
		return value
	}
	return nil
}

func (app *server) error404(req *http.Request) ActionResult {
	res := NewActionResult()
	res.SetStatusCode(404)
	res.Write(renderError(404, "Not Found", "The resource you are looking for has been removed, had its name changed, or is temporarily unavailable", ""))
	return res
}

func (app *server) error403(req *http.Request) ActionResult {
	res := NewActionResult()
	res.SetStatusCode(403)
	res.Write(renderError(403, "Forbidden", `The server understood the request but refuses to authorize it: <b>` + req.URL.Path + `</b>`, ""))
	return res
}

func (app *server) handleError(req *http.Request, code int) ActionResult {
	var handler = app.errorHandlers[code]
	if handler != nil {
		return handler(req)
	}
	app.logWriter().Fatalln("unhandled request", req.Method, "'"+req.URL.Path+"'")
	return app.error404(req)
}

func (app *server) viewFolder() string {
	return app.MapPath("/views")
}

func (app *server) logWriter() *log.Logger {
	if app.logger == nil {
		app.logger = log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)
	}
	return app.logger
}

func (app *server) panicRecover(res http.ResponseWriter, req *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	// detect end request
	_,ok := rec.(*endRequestError)
	if ok {
		return
	}
	// process 500 error
	res.WriteHeader(500)
	if err, ok := rec.(error); ok {
		var debugStack = string(debug.Stack())
		debugStack = strings.Replace(debugStack, "<", "&lt;", -1)
		debugStack = strings.Replace(debugStack, ">", "&gt;", -1)
		res.Write(renderError(500, "Internal Server Error", err.Error(), debugStack))
	} else {
		res.Write(renderError(500, "Internal Server Error", err.Error(), ""))
	}
}

func newServer(webRoot string) *server {
	var app = &server{
		webRoot:       webRoot,
		routeLocked:   false,
		errorHandlers: make(map[int]CtxHandler),
	}
	app.views = make(map[string]*view)
	app.filters = make(map[string][]FilterFunc)
	app.viewExt = ".html"
	return app
}