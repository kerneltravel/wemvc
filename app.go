package wemvc

import (
	"net/http"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) *Result

// RootDir get the root file path of the web server
func RootDir() string {
	return svr.RootDir()
}

// Config get the config data
func Config() Configuration {
	return svr.Config()
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return svr.MapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NsSection {
	return svr.Namespace(ns)
}

// AddViewFunc add the view func map
func AddViewFunc(name string, f interface{}) {
	svr.AddViewFunc(name, f)
}

// AddRouteFunc add the route analyze helper function to server
func AddRouteFunc(name string, f RouteValidateFunc) {
	svr.AddRouteFunc(name, f)
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) {
	svr.SetRootDir(rootDir)
}

// SetViewExt set the view file extension
func SetViewExt(ext string) {
	svr.SetViewExt(ext)
}

// StaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) {
	svr.StaticDir(pathPrefix)
}

// StaticFile serve the path as static file
func StaticFile(path string) {
	svr.StaticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler CtxHandler) {
	svr.HandleError(errorCode, handler)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) {
	svr.Route(routePath, c, defaultAction...)
}

// Filter set the route filter
func Filter(pathPrefix string, filter CtxFilter) {
	svr.SetPathFilter(pathPrefix, filter)
}

func SetGlobalFilters(filters []CtxFilter) {
	svr.SetGlobalFilter(filters)
}

// Run run the web application
func Run(port int) {
	svr.Run(port)
}

// Run run the web application
func RunTLS(port int, certFile, keyFile string) {
	svr.RunTLS(port, certFile, keyFile)
}

// App the application singleton
var (
	svr = newServer(WorkingDir())
)
