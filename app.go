package wemvc

import (
	"log"
	"net/http"
	"os"
	"sync"
)

// CtxHandler the error handler define
type CtxHandler func(*http.Request) Result

// FilterFunc request filter func
type FilterFunc func(ctx Context)

// RootDir get the root file path of the web server
func RootDir() string {
	return DefaultServer.RootDir()
}

// Config get the config data
func Config() Configuration {
	return DefaultServer.Config()
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return DefaultServer.MapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NamespaceSection {
	return DefaultServer.Namespace(ns)
}

// AddViewFunc add the view func map
func AddViewFunc(name string, f interface{}) Server {
	return DefaultServer.AddViewFunc(name, f)
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) Server {
	return DefaultServer.SetRootDir(rootDir)
}

// SetViewExt set the view file extension
func SetViewExt(ext string) Server {
	return DefaultServer.SetViewExt(ext)
}

// ServeStaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) Server {
	return DefaultServer.StaticDir(pathPrefix)
}

// ServeStaticFile serve the path as static file
func StaticFile(path string) Server {
	return DefaultServer.StaticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler CtxHandler) Server {
	return DefaultServer.HandleError(errorCode, handler)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) Server {
	return DefaultServer.Route(routePath, c, defaultAction...)
}

// SetFilter set the route filter
func Filter(pathPrefix string, filter FilterFunc) Server {
	return DefaultServer.Filter(pathPrefix, filter)
}

// Logger return the log writer
func Logger() *log.Logger {
	return DefaultServer.Logger()
}

// SetLogFile set the log file, the default log file is os.Stdout
func SetLogFile(name string) Server {
	return DefaultServer.SetLogFile(name)
}

// Run run the web application
func Run(port int) {
	DefaultServer.Run(port)
}

func NewServer(webRoot string) Server {
	return newServer(webRoot)
}

func WaitForExit() {
	serverWaiting.Wait()
}

// App the application singleton
var (
	DefaultServer = newServer(WorkingDir())
	serverWaiting = sync.WaitGroup{}
)

func WorkingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
