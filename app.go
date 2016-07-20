package wemvc

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strings"
	"sort"
	"github.com/Simbory/wemvc/fsnotify"
	"github.com/Simbory/wemvc/session"
	"github.com/Simbory/wemvc/utils"
)

// Handler the error handler define
type Handler func(*http.Request) ActionResult

type Filter func(ctx Context)

type appServer struct {
	errorHandlers map[int]Handler
	Port           int
	webRoot        string
	config         *configuration
	router         *Router
	watcher        *fsnotify.Watcher
	watchingFiles  []string
	initError      error
	routeLocked    bool
	staticPaths    []string
	filters        map[string][]Filter
	globalSession  *session.SessionManager
}

// GetRootPath get the root file path of the web server
func (app *appServer) GetRootPath() string {
	return app.webRoot
}

// GetConfig get the config data
func (app *appServer) GetConfig() Configuration {
	return app.config
}

// MapPath Returns the physical file path that corresponds to the specified virtual path.
func (app *appServer) MapPath(virtualPath string) string {
	var res = path.Join(app.GetRootPath(), virtualPath)
	return utils.FixPath(res)
}

// SetStaticPath set the path as a static path that the file under this path is served as static file
func (app *appServer) SetStaticPath(path string) {
	if len(path) < 1 {
		panic(errors.New("the static path prefix cannot be empty"))
	}
	if !strings.HasPrefix(path, "/") {
		panic(errors.New("The static path prefix should start with '/'"))
	}
	var fullPath = app.MapPath(path)
	if utils.IsDir(fullPath) && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	app.staticPaths = append(app.staticPaths, strings.ToLower(path))
}

func (app *appServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// check init error
	if app.initError != nil {
		w.WriteHeader(500)
		w.Write([]byte(app.initError.Error()))
		return
	}
	defer app.panicRecover(w, req)

	var lUrl = strings.ToLower(req.URL.Path)
	var ctx = &context{
		req: req,
		w:   w,
		end: false,
	}
	// execute the filters
	var tmpFilters = app.filters
	var keys []string
	for key := range tmpFilters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if ctx.end {
			return
		}
		if strings.HasPrefix(lUrl+"/", key) {
			for _, f := range tmpFilters[key] {
				f(ctx)
			}
		}
	}
	if ctx.end {
		return
	}

	// serve the static file
	for _, p := range app.staticPaths {
		if strings.HasPrefix(lUrl, p) {
			app.serveStaticFile(w, req)
			return
		}
	}

	// serve the dynamic page
	var result ActionResult
	result = app.serveDynamic(ctx)
	// handle error 404
	if result == nil {
		result = app.showError(req, 404)
	}
	// process the dynamic result
	res, ok := result.(*actionResult)
	if ok {
		if len(res.resFile) > 0 {
			http.ServeFile(w, req, res.resFile)
			return
		}
		if len(res.redUrl) > 0 {
			http.Redirect(w, req, res.redUrl, res.statusCode)
			return
		}
		// write the result to browser
		for k, v := range result.GetHeaders() {
			//fmt.Println("Key: ", k, " Value: ", v)
			w.Header().Add(k, v)
		}
		var contentType = fmt.Sprintf("%s;charset=%s", result.GetContentType(), result.GetEncoding())
		w.Header().Add("Content-Type", contentType)
		if result.GetStatusCode() != 200 {
			w.WriteHeader(result.GetStatusCode())
		}
		var output = result.GetOutput()
		if len(output) > 0 {
			w.Write(result.GetOutput())
		}
	}
}

// SetErrorHandler set the error handler
func (app *appServer) SetErrorHandler(code int, handler Handler) {
	if app.errorHandlers == nil {
		app.errorHandlers = make(map[int]Handler)
	}
	app.errorHandlers[code] = handler
}

// Route set the route
func (app *appServer) Route(strPth string, c interface{}) {
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

// SetFilter add filter to each request
func (app *appServer) SetFilter(pathPrefix string, filter Filter) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic("the filter path preix must starts with \"/\"")
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	app.filters[strings.ToLower(pathPrefix)] = append(app.filters[strings.ToLower(pathPrefix)], filter)
}

func (app *appServer) Run() error {
	app.routeLocked = true
	println("root:", app.webRoot)
	println("port:", app.Port)
	port := fmt.Sprintf(":%d", app.Port)
	err := http.ListenAndServe(port, app)
	return err
}

// App the application singleton
var AppServer *appServer

func init() {
	app, err := newApp(8080)
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
	AppServer = app
}

func getWorkPath() string {
	p,err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

func newApp(port int) (*appServer, error) {
	var root = getWorkPath()
	app := &appServer{
		webRoot:     root,
		Port:        port,
		initError:   nil,
		routeLocked: false,
		filters:     make(map[string][]Filter),
	}
	err := app.init()
	return app, err
}
