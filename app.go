package wemvc

import (
	"net/http"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) *Result

// RootDir get the root file path of the web server
func RootDir() string {
	return _server.RootDir()
}

// Config get the config data
func Config() Configuration {
	return _server.Config()
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return _server.MapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NsSection {
	return _server.Namespace(ns)
}

// AddViewFunc add the view func map
func AddViewFunc(name string, f interface{}) {
	_server.AddViewFunc(name, f)
}

// AddRouteFunc add the route analyze helper function to server
func AddRouteFunc(name string, f RouteValidateFunc) {
	_server.AddRouteFunc(name, f)
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) {
	_server.SetRootDir(rootDir)
}

// SetViewExt set the view file extension
func SetViewExt(ext string) {
	_server.SetViewExt(ext)
}

// Filter set the route filter
func SetPathFilter(pathPrefix string, filter CtxFilter) {
	_server.SetPathFilter(pathPrefix, filter)
}

func SetGlobalFilters(filters []CtxFilter) {
	_server.SetGlobalFilter(filters)
}

// StaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) {
	_server.StaticDir(pathPrefix)
}

// StaticFile serve the path as static file
func StaticFile(path string) {
	_server.StaticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler CtxHandler) {
	_server.HandleError(errorCode, handler)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) {
	_server.Route(routePath, c, defaultAction...)
}

// PrintRouteInfo print route tree information
func PrintRouteInfo() []byte {
	return _server.printRoute()
}

// RenderView render the view template and get the result
func RenderView(viewName string, data interface{}) ([]byte, error) {
	return _server.renderView(viewName, data)
}

// Run run the web application
func Run(port int) {
	_server.Run(port)
}

// Run run the web application
func RunTLS(port int, certFile, keyFile string) {
	_server.RunTLS(port, certFile, keyFile)
}

// App the application singleton
var (
	_server = newServer(WorkingDir())
)
