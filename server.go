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

	"container/list"
	"net/url"
	"runtime"
	"time"
)

type EventHandler func() error

type server struct {
	errorHandlers   map[int]ErrorHandler
	domain          string
	port            int
	webRoot         string
	config          *config
	routing         *routeTree
	locked          bool
	staticPaths     []string
	staticFiles     []string
	globalSession   *SessionManager
	namespaces      map[string]*NsSection
	sessionProvides map[string]SessionProvider
	internalErr     error
	fileWatcher     *FileWatcher
	cacheManager    *CacheManager
	appInitEvents   []EventHandler
	httpReqEvents   map[requestEvent][]CtxFilter
	viewContainer
	filterContainer
}

func (app *server) onAppInit(h EventHandler) {
	app.assertNotLocked()
	if h == nil {
		return
	}
	app.appInitEvents = append(app.appInitEvents, h)
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
func (app *server) mapPath(virtualPath string) string {
	var res = path.Join(app.webRoot, virtualPath)
	return fixPath(res)
}

func (app *server) RegSessionProvider(name string, provide SessionProvider) {
	app.assertNotLocked()
	if provide == nil {
		panic(errSessionProvNil)
	}
	if _, dup := app.sessionProvides[name]; dup {
		panic(errSessionRegTwice(name))
	}
	app.sessionProvides[name] = provide
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
	app.execReqEvents(beforeCheck, ctx)
	app.execReqEvents(afterCheck, ctx)
	if app.isStaticRequest(req) {
		app.execReqEvents(beforeStatic, ctx)
		app.execReqEvents(afterStatic, ctx)
	}
	app.execReqEvents(beforeRoute, ctx)
	app.execReqEvents(afterRoute, ctx)
	if !ctx.ended {
		execFilters(ctx)
	}
	app.execReqEvents(beforeAction, ctx)
	app.execReqEvents(afterAction, ctx)
	// flush the request
	app.flushRequest(w, req, ctx.Result)
}

func (app *server) staticDir(pathPrefix string) {
	app.assertNotLocked()
	if len(pathPrefix) < 1 {
		panic(errPathPrefix)
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = strAdd("/", pathPrefix)
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = strAdd(pathPrefix, "/")
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
		path = strAdd("/", path)
	}
	if strings.HasSuffix(path, "/") {
		panic(errInvalidPath)
	}
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}
	app.staticFiles = append(app.staticFiles, path)
}

func (app *server) getNamespace(nsName string) *NsSection {
	if len(nsName) > 0 {
		if !strings.HasPrefix(nsName, "/") {
			nsName = strAdd("/", nsName)
		}
		nsName = strings.TrimRight(nsName, "/")
	}
	if len(nsName) < 1 {
		panic(errInvalidNamespace)
	}
	if app.namespaces == nil {
		app.namespaces = make(map[string]*NsSection)
	}
	ns, ok := app.namespaces[nsName]
	if ok {
		return ns
	}
	ns = &NsSection{
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
	cInfo := newControllerInfo(namespace, t, action)
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
	case Result:
		result.(Result).ExecResult(w, req)
		return
	case *url.URL:
		res := result.(*url.URL)
		http.Redirect(w, req, res.String(), 302)
		return
	case string:
		content := result.(string)
		w.Header().Add("Content-Type", "text/plain")
		w.Write(str2Byte(content))
		return
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

func (app *server) initWatcher() error {
	// add config file handler
	app.fileWatcher.AddHandler(&configDetector{app: app})
	// add ns config handler
	app.fileWatcher.AddHandler(&nsConfigDetector{app: app})
	// add view file handler
	app.fileWatcher.AddHandler(&viewDetector{app: app})
	// add ns view file handler
	app.fileWatcher.AddHandler(&nsViewDetector{app: app})
	// start file watcher
	app.fileWatcher.Start()
	return nil
}

func (app *server) initConfig() error {
	// load & watch the global config files
	globalConfigFile := app.mapPath("/config.xml")
	conf, err := newConfig(globalConfigFile)
	if err != nil {
		return err
	}
	if conf == nil {
		conf = &config{
			defaultUrls: []string{"index.html", "index.htm"},
			DefaultURL:  "index.html;index.htm",
			SessionConfig: &SessionConfig{
				ManagerName:     "memory",
				CookieName:      "Session_ID",
				EnableSetCookie: true,
				SessionIDLength: 32,
			},
		}
	} else {
		err1 := app.fileWatcher.AddWatch(globalConfigFile)
		if err1 != nil {
			return err
		}
	}
	app.config = conf
	return nil
}

func (app *server) initViews() error {
	app.addViewFunc("include", include_view)
	app.addViewFunc("req_query", req_query)
	app.addViewFunc("req_form", req_form)
	app.addViewFunc("req_header", req_header)
	app.addViewFunc("req_postForm", req_postForm)
	app.addViewFunc("req_host", req_host)
	app.addViewFunc("cache", cache_view)
	app.addViewFunc("session", session_view)
	// build the view template and watch the changes
	viewDir := app.viewFolder()
	if IsDir(viewDir) {
		app.compileViews(viewDir)
		if app.fileWatcher != nil {
			app.fileWatcher.AddWatch(viewDir)
			filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error {
				if info.IsDir() {
					app.fileWatcher.AddWatch(p)
				}
				return nil
			})
		}
	}
	return nil
}

func (app *server) initNs() error {
	// process namespaces: build the views files and load the config
	if app.namespaces != nil {
		for _, ns := range app.namespaces {
			//app.logWriter().Println("process namespace", name)
			settingFile := ns.nsSettingFile()
			ns.loadConfig()
			if app.fileWatcher != nil {
				app.fileWatcher.AddWatch(settingFile)
			}
			nsViewDir := ns.viewFolder()
			ns.compileViews(nsViewDir)
			if app.fileWatcher != nil {
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
	}
	return nil
}

func (app *server) initSessionMgr() error {
	// init sessionManager
	app.RegSessionProvider("memory", &memSessionProvider{list: list.New(), sessions: make(map[string]*list.Element)})
	mgr, err := app.NewSessionManager(app.config.SessionConfig.ManagerName, app.config.SessionConfig)
	if err != nil {
		return err
	}
	app.globalSession = mgr
	go app.globalSession.GC()
	return nil
}

func (app *server) initCacheMgr() error {
	// init cache manager
	app.cacheManager = newCacheManager(app.fileWatcher, 10*time.Second)
	app.cacheManager.start()
	return nil
}

func (app *server) initErrorHandler() error {
	app.errorHandlers[404] = app.error404
	app.errorHandlers[403] = app.error403
	return nil
}

func (app *server) init() error {
	// init the error handler
	var err error
	for _, h := range app.appInitEvents {
		err = h()
		if err != nil {
			return err
		}
	}
	return nil
}

func (app *server) isConfigFile(f string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(app.mapPath("/config.xml"), f)
	} else {
		return app.mapPath("/config.xml") == f
	}
}

func (app *server) isInViewFolder(f string) bool {
	viewPath := app.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

// isStaticRequest check the current request is indicate to static path
func (app *server) isStaticRequest(req *http.Request) bool {
	var reqUrl string
	if runtime.GOOS == "windows" {
		reqUrl = strings.ToLower(req.URL.Path)
	} else {
		reqUrl = req.URL.Path
	}
	for _, f := range app.staticFiles {
		if f == reqUrl {
			return true
		}
	}
	for _, p := range app.staticPaths {
		if strings.HasPrefix(reqUrl, p) {
			return true
		}
	}
	return false
}

func (app *server) viewFolder() string {
	return app.mapPath("/views")
}

// regRequestFilter register context filter to the featured request step
func (app *server) regRequestFilter(ev requestEvent, h CtxFilter) {
	app.assertNotLocked()
	hs, ok := app.httpReqEvents[ev]
	if ok && h != nil {
		hs = append(hs, h)
		app.httpReqEvents[ev] = hs
	}
}

func (app *server) execReqEvents(ev requestEvent, ctx *Context) {
	if ctx == nil || ctx.ended {
		return
	}
	events, ok := app.httpReqEvents[ev]
	if ok && len(events) > 0 {
		for _, h := range events {
			if h(ctx); ctx.ended {
				return
			}
		}
	}
}

func newServer(webRoot string) *server {
	var app = &server{
		webRoot:       webRoot,
		locked:        false,
		errorHandlers: make(map[int]ErrorHandler),
	}
	app.views = make(map[string]*view)
	app.filters = make(map[string][]CtxFilter)
	app.viewExt = ".html"
	app.sessionProvides = make(map[string]SessionProvider)
	app.httpReqEvents = make(map[requestEvent][]CtxFilter, 8)
	app.httpReqEvents[beforeCheck] = nil
	app.httpReqEvents[afterCheck] = []CtxFilter{dangerCheck}
	app.httpReqEvents[beforeStatic] = nil
	app.httpReqEvents[afterStatic] = []CtxFilter{serveStatic}
	app.httpReqEvents[beforeRoute] = nil
	app.httpReqEvents[afterRoute] = []CtxFilter{handleRoute}
	app.httpReqEvents[beforeAction] = nil
	app.httpReqEvents[afterAction] = []CtxFilter{execAction}
	app.onAppInit(app.initWatcher)
	app.onAppInit(app.initConfig)
	app.onAppInit(app.initViews)
	app.onAppInit(app.initNs)
	app.onAppInit(app.initSessionMgr)
	app.onAppInit(app.initCacheMgr)
	w, err := NewWatcher()
	if err != nil {
		panic(err)
	}
	app.fileWatcher = w
	return app
}
