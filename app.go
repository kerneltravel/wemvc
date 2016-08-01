package wemvc

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"

	"log"

	"github.com/Simbory/wemvc/utils"
)

// Handler the error handler define
type Handler func(*http.Request) ActionResult

// Filter request filter func
type Filter func(ctx Context)

// SetRootDir set the webroot of the web application
func SetRootDir(rootDir string) {
	if app.routeLocked {
		panic(errors.New("Cannot set the web root while the application is running."))
	}
	if !utils.IsDir(rootDir) {
		panic("invalid root dir")
	}
	app.webRoot = rootDir
}

// RootDir get the root file path of the web server
func RootDir() string {
	return app.webRoot
}

// Config get the config data
func Config() Configuration {
	return app.config
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
// @param virtualPath: the virtual path starts with
// @return the absolute file path
func MapPath(virtualPath string) string {
	return app.mapPath(virtualPath)
}

// AddStatic set the path as a static path that the file under this path is served as static file
// @param pathPrefix: the path prefix starts with '/'
func AddStatic(pathPrefix string) {
	if len(pathPrefix) < 1 {
		panic(errors.New("the static path prefix cannot be empty"))
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		panic(errors.New("The static path prefix should start with '/'"))
	}
	var fullPath = MapPath(pathPrefix)
	if utils.IsDir(fullPath) && !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	app.staticPaths = append(app.staticPaths, strings.ToLower(pathPrefix))
}

// HandleError handle the error code with the error handler
func HandleError(errorCode int, handler Handler) {
	app.errorHandlers[errorCode] = handler
}

// Route set the route
func Route(routePath string, c interface{}, defaultAction ...string) {
	if app.routeLocked {
		println("The controller cannot be added after the application is started.")
		os.Exit(-1)
	}
	var t = reflect.TypeOf(c)
	var action = "index"
	if len(defaultAction) > 0 && len(defaultAction[0]) > 0 {
		action = defaultAction[0]
	}
	cInfo := newControllerInfo(t, action)
	if app.router == nil {
		app.router = newRouter()
	}
	app.logWriter().Println("set route '"+routePath+"'        controller:", cInfo.controllerType.Name(), "       default action:", cInfo.defaultAction+"\r\n")
	app.router.Handle(routePath, cInfo)
}

// SetFilter set the route filter
func SetFilter(pathPrefix string, filter Filter) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic("the filter path preix must starts with \"/\"")
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	app.filters[strings.ToLower(pathPrefix)] = append(app.filters[strings.ToLower(pathPrefix)], filter)
}

func Logger() *log.Logger {
	return app.logWriter()
}

// SetLogFile set the log file, the default log file is os.Stdout
func SetLogFile(name string) {
	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	logger := log.New(file, "", log.LstdFlags|log.Llongfile)
	app.logger = logger
}

// Run run the web application
func Run(port int) error {
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

// App the application singleton
var app *server

func init() {
	var root = getWorkPath()
	app = &server{
		webRoot:     	root,
		initError:   	nil,
		routeLocked: 	false,
		filters:     	make(map[string][]Filter),
		views:       	make(map[string]*view),
		errorHandlers:	make(map[int]Handler),
	}
}

func getWorkPath() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}
