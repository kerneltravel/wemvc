package wemvc

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"github.com/Simbory/wemvc/utils"
)

// Handler the error handler define
type Handler func(*http.Request) ActionResult

type Filter func(ctx Context)

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
	var res = path.Join(RootDir(), virtualPath)
	return utils.FixPath(res)
}

/* AddStatic set the path as a static path that the file under this path is served as static file
 * @param pathPrefix: the path prefix starts with '/'
 */
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

// HandleErr handle the error code with the error handler
func HandleErr(errorCode int, handler Handler) {
	if app.errorHandlers == nil {
		app.errorHandlers = make(map[int]Handler)
	}
	app.errorHandlers[errorCode] = handler
}

// Route set the route
func Route(strPth string, c interface{}) {
	if app.routeLocked {
		println("The controller cannot be added after the application is started.")
		os.Exit(-1)
	}
	var t = reflect.TypeOf(c)
	cInfo := createControllerInfo(t)
	if app.router == nil {
		app.router = newRouter()
	}
	app.router.Handle(strPth, cInfo)
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

// Run run the web application
func Run(port int) error {
	app.routeLocked = true
	app.port = port
	println("root:", app.webRoot)
	println("port:", app.port)
	portStr := fmt.Sprintf(":%d", app.port)
	return http.ListenAndServe(portStr, app)
}

// App the application singleton
var app *server

func init() {
	var root = getWorkPath()
	app = &server{
		webRoot:     root,
		initError:   nil,
		routeLocked: false,
		filters:     make(map[string][]Filter),
	}
	err := app.init()
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
}

func getWorkPath() string {
	p,err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}