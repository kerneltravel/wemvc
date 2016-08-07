package wemvc

import (
	"log"
	"net/http"
	"os"
)

// Handler the error handler define
type CtxHandler func(*http.Request) ActionResult

// Filter request filter func
type FilterFunc func(ctx Context)

// RootDir get the root file path of the web server
func RootDir() string {
	return app.RootDir()
}

// Config get the config data
func Config() Configuration {
	return app.Config()
}

// Application the application interface that define the useful function
type Application interface {
	RootDir() string
	Config() Configuration
	MapPath(virtualPath string) string
	Logger() *log.Logger
	Namespace(ns string) NamespaceSection
	SetRootDir(rootDir string) Application
	StaticDir(pathPrefix string) Application
	StaticFile(path string) Application
	HandleError(errorCode int, handler CtxHandler) Application
	Route(routePath string, c interface{}, defaultAction ...string) Application
	Filter(pathPrefix string, filter FilterFunc) Application
	SetLogFile(name string) Application
	AddViewFunc(name string, f interface{}) Application
	Run(port int) error
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return app.MapPath(virtualPath)
}

// Namespace return the namespace by name
func Namespace(ns string) NamespaceSection {
	return app.Namespace(ns)
}

func AddViewFunc(name string, f interface{}) Application {
	return app.AddViewFunc(name, f)
}

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) Application {
	return app.SetRootDir(rootDir)
}

// ServeStaticDir set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func StaticDir(pathPrefix string) Application {
	return app.StaticDir(pathPrefix)
}

// ServeStaticFile serve the path as static file
func StaticFile(path string) Application {
	return app.StaticFile(path)
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler CtxHandler) Application {
	return app.HandleError(errorCode, handler)
}

// Route set the route rule
func Route(routePath string, c interface{}, defaultAction ...string) Application {
	return app.Route(routePath, c, defaultAction...)
}

// SetFilter set the route filter
func Filter(pathPrefix string, filter FilterFunc) Application {
	return app.Filter(pathPrefix, filter)
}

// Logger return the log writer
func Logger() *log.Logger {
	return app.Logger()
}

// SetLogFile set the log file, the default log file is os.Stdout
func SetLogFile(name string) Application {
	return app.SetLogFile(name)
}

// Run run the web application
func Run(port int) error {
	return app.Run(port)
}

// App the application singleton
var app *server

func init() {
	var root = getWorkPath()
	app = &server{
		webRoot:       root,
		routeLocked:   false,
		errorHandlers: make(map[int]CtxHandler),
	}
	app.views = make(map[string]*view)
	app.filters = make(map[string][]FilterFunc)
}

func getWorkPath() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
