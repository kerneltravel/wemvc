package wemvc

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) *ContentResult

// RootDir get the root file path of the web server
func RootDir() string {
	return app.webRoot
}

// Config get the config data
func Config() Configuration {
	return app.config
}

func Cache() *CacheManager {
	return app.cacheManager
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return app.mapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NsSection {
	return app.getNamespace(ns)
}

// AddViewFunc add the view func map
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
func SetPathFilter(pathPrefix string, filter CtxFilter) {
	app.assertNotLocked()
	if !app.routing.MatchCase {
		pathPrefix = strings.ToLower(pathPrefix)
	}
	app.setFilter(pathPrefix, filter)
}

func SetGlobalFilters(filters []CtxFilter) {
	app.assertNotLocked()
	if len(filters) < 1 {
		return
	}
	app.globalFilters = filters
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
func HandleError(errorCode int, handler CtxHandler) {
	app.assertNotLocked()
	app.errorHandlers[errorCode] = handler
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

// Run run the web application
func Run(port int) {
	err := app.init()
	if err != nil {
		panic(err)
	}
	app.locked = true
	app.port = port
	portStr := fmt.Sprintf(":%d", app.port)
	err = http.ListenAndServe(portStr, app)
	if err != nil {
		panic(err)
	}
}

// Run run the web application
func RunTLS(port int, certFile, keyFile string) {
	err := app.init()
	if err != nil {
		panic(err)
	}
	app.locked = true
	app.port = port
	portStr := fmt.Sprintf(":%d", app.port)
	err = http.ListenAndServeTLS(portStr, certFile, keyFile, app)
	if err != nil {
		panic(err)
	}
}

var app *server

func init() {
	app = newServer(WorkingDir())
}
