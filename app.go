package wemvc

import (
	//"log"
	"net/http"
	"os"
	"sync"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) *Result

// FilterFunc request filter func
type FilterFunc func(ctx Context)

// RootDir get the root file path of the web server
func RootDir() string {
	return defaultServer.RootDir()
}

// Config get the config data
func Config() Configuration {
	return defaultServer.Config()
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return defaultServer.MapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NsSection {
	return defaultServer.Namespace(ns)
}

// AddViewFunc add the view func map
func AddViewFunc(name string, f interface{}) {
	defaultServer.AddViewFunc(name, f)
}

// AddRouteFunc add the route analyze helper function to server
func AddRouteFunc(name string, f RouteValidateFunc) {
	defaultServer.AddRouteFunc(name, f)
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) {
	defaultServer.SetRootDir(rootDir)
}

// SetViewExt set the view file extension
func SetViewExt(ext string) {
	defaultServer.SetViewExt(ext)
}

// StaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) {
	defaultServer.StaticDir(pathPrefix)
}

// StaticFile serve the path as static file
func StaticFile(path string) {
	defaultServer.StaticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler CtxHandler) {
	defaultServer.HandleError(errorCode, handler)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) {
	defaultServer.Route(routePath, c, defaultAction...)
}

// Filter set the route filter
func Filter(pathPrefix string, filter FilterFunc) {
	defaultServer.Filter(pathPrefix, filter)
}

// Run run the web application
func Run(port int) {
	defaultServer.Run(port)
}

// Run run the web application
func RunTLS(port int, certFile, keyFile string) {
	defaultServer.RunTLS(port, certFile, keyFile)
}

// App the application singleton
var (
	defaultServer = newServer(WorkingDir())
)
