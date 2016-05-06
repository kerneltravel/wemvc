package wemvc

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

type application struct {
	errorHandlers map[int]Handler
	port           int
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

func (app *application) GetWebRoot() string {
	return app.webRoot
}

func (app *application) Port(p ...int) int {
	if len(p) > 0 {
		app.port = p[0]
	}
	return app.port
}

func (app *application) GetConfig() Configuration {
	return app.config
}

func (app *application) MapPath(relativePath string) string {
	var res = path.Join(app.GetWebRoot(), relativePath)
	return utils.FixPath(res)
}

func (app *application) InitSessionManager(name string) {
}

func (app *application) SetStaticPath(path string) {
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

func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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

func (app *application) SetErrorHandler(code int, handler Handler) {
	if app.errorHandlers == nil {
		app.errorHandlers = make(map[int]Handler)
	}
	app.errorHandlers[code] = handler
}

func (app *application) Route(strPth string, c interface{}) {
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
func (app *application) SetFilter(pathPrefix string, filter Filter) {
	if !strings.HasPrefix(pathPrefix, "") {
		panic("the filter path preix must starts with \"/\"")
	}
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	app.filters[strings.ToLower(pathPrefix)] = append(app.filters[strings.ToLower(pathPrefix)], filter)
}

func (app *application) Run() error {
	app.routeLocked = true
	port := fmt.Sprintf(":%d", app.port)
	println("website started")
	err := http.ListenAndServe(port, app)
	return err
}

// App the application singleton
var App *application

func init() {
	root := flag.String("root", "", "the server root")
	port := flag.Int("port", 8080, "server running port")
	flag.Parse()
	var appPort = *port
	var appRoot = *root
	if len(appRoot) < 1 {
		println("arguments:")
		flag.PrintDefaults()
		appRoot = utils.GetCurrentDirectory()
	}
	println("using root:", appRoot)
	println("using port:", appPort)
	app, err := newApp(appRoot, appPort)
	if err != nil {
		println(err.Error())
		os.Exit(-1)
	}
	App = app
}

func newApp(root string, port int) (*application, error) {
	if len(root) < 1 {
		return nil, errors.New("Web root cannot be empty.")
	}
	webRoot := strings.TrimSuffix(strings.TrimSuffix(root, "\\"), "/")
	if strings.HasPrefix(webRoot, ".") {
		file, _ := exec.LookPath(os.Args[0])
		exePath, _ := filepath.Abs(file)
		exeDir := filepath.Dir(exePath)
		webRoot = path.Join(exeDir, webRoot)
	}

	if !utils.IsDir(webRoot) {
		return nil, errors.New("Path \"" + webRoot + "\" is not a directory")
	}
	app := &application{
		webRoot:     utils.FixPath(webRoot),
		port:        port,
		initError:   nil,
		routeLocked: false,
		filters:     make(map[string][]Filter),
	}
	err := app.init()
	return app, err
}
