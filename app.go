package wemvc

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// CtxHandler the error handler define
type ErrorHandler func(*http.Request) *ContentResult

// RootDir get the root file path of the web server
func RootDir() string {
	return app.webRoot
}

// Config get the config data
func Config() Configuration {
	return app.config
}

// Cache get the cache manager
func Cache() *CacheManager {
	return app.cacheManager
}

// Watcher get the file watcher
func Watcher() *FileWatcher {
	return app.fileWatcher
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return app.mapPath(virtualPath)
}

// Namespace get the namespace by name
func Namespace(ns string) *NsSection {
	return app.getNamespace(ns)
}

// AddViewFunc add the view func to the view func map
func AddViewFunc(name string, f interface{}) {
	app.assertNotLocked()
	app.addViewFunc(name, f)
}

// AddRouteFunc add the route analyze helper function to server
func AddRouteFunc(name string, f RouteValidateFunc) {
	app.assertNotLocked()
	err := app.routing.addFunc(name, f)
	if err != nil {
		panic(err)
	}
}

// SetDomain set the server domain
func SetDomain(domain string) {
	app.domain = domain
}

// GetDomain get the server domain
func GetDomain() string {
	return app.domain
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) {
	app.assertNotLocked()
	if !IsDir(rootDir) {
		panic(errInvalidRoot)
	}
	app.webRoot = rootDir
}

// SetViewExt set the view file extension
func SetViewExt(ext string) {
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
			ns.viewExt = ext
		}
	}
}

// SetPathFilter set the route path filter
func SetPathFilter(pathPrefix string, filterFunc CtxFilter) {
	app.assertNotLocked()
	if !app.routing.MatchCase {
		pathPrefix = strings.ToLower(pathPrefix)
	}
	app.addFilter(pathPrefix, filterFunc)
}

// StaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) {
	app.staticDir(pathPrefix)
}

// StaticFile serve the path as static file
func StaticFile(path string) {
	app.staticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler ErrorHandler) {
	app.assertNotLocked()
	app.errorHandlers[errorCode] = handler
}

// OnAppInit add app init handler
func OnAppInit(h EventHandler) {
	app.onAppInit(h)
}

// BeforeCheck register context filter to before security check step
func BeforeCheck(filterFunc CtxFilter) {
	app.regRequestFilter(beforeCheck, filterFunc)
}

// AfterCheck register context filter to after security check step
func AfterCheck(filterFunc CtxFilter) {
	app.regRequestFilter(afterCheck, filterFunc)
}

// BeforeServeStatic register context filter to before serving static file step
func BeforeServeStatic(filterFunc CtxFilter) {
	app.regRequestFilter(beforeStatic, filterFunc)
}

// AfterServeStatic register context filter to after serving static file step
func AfterServeStatic(filterFunc CtxFilter) {
	app.regRequestFilter(afterStatic, filterFunc)
}

// BeforeRoute register context filter to before routing step
func BeforeRoute(filterFunc CtxFilter) {
	app.regRequestFilter(beforeRoute, filterFunc)
}

// AfterRoute register context filter to after routing step
func AfterRoute(filterFunc CtxFilter) {
	app.regRequestFilter(afterRoute, filterFunc)
}

// BeforeExecAction register context filter to before executing action step
func BeforeExecAction(filterFunc CtxFilter) {
	app.regRequestFilter(beforeAction, filterFunc)
}

// AfterExecAction register context filter to after executing action step
func AfterExecAction(filterFunc CtxFilter) {
	app.regRequestFilter(afterAction, filterFunc)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) {
	app.assertNotLocked()
	action := "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	app.addRoute("", routePath, c, action)
}

// PrintRouteInfo print route tree information
func PrintRouteInfo() []byte {
	return data2Json(app.routing)
}

// RenderView render the view template and get the result
func RenderView(viewName string, data interface{}) ([]byte, error) {
	return app.renderView(viewName, data)
}

// RegSessionProvider register session provider
func RegSessionProvider(name string, provider SessionProvider) {
	app.regSessionProvider(name, provider)
}

// Run run the web application
func Run(port int) {
	err := app.init()
	if err != nil {
		panic(err)
	}
	app.locked = true
	app.port = port
	portStr := fmt.Sprintf("%s:%d", app.domain, app.port)
	err = http.ListenAndServe(portStr, app)
	if err != nil {
		panic(err)
	}
}

// RunTLS run the web application as TLS
func RunTLS(port int, certFile, keyFile string) {
	err := app.init()
	if err != nil {
		panic(err)
	}
	app.locked = true
	app.port = port
	portStr := fmt.Sprintf("%s:%d", app.domain, app.port)
	err = http.ListenAndServeTLS(portStr, certFile, keyFile, app)
	if err != nil {
		panic(err)
	}
}

var app *server

func init() {
	app = newServer(WorkingDir())
}
