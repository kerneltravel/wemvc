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

	"github.com/Simbory/wemvc/fsnotify"
)

// Handler the error handler define
type Handler func(*http.Request) Response

type application struct {
	errorHandlers map[int]Handler
	port          int
	webRoot       string
	config        *configuration
	router        *Router
	watcher       *fsnotify.Watcher
	viewsWatcher  *fsnotify.Watcher
	watchingFiles []string
	initError     error
	routeLocked   bool
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
	return fixPath(res)
}

func (app *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// check init error
	if app.initError != nil {
		w.WriteHeader(500)
		w.Write([]byte(app.initError.Error()))
		return
	}
	//defer app.panicRecover(w, req)

	// serve the dynamic page
	var result Response
	result = app.serveDynamic(w, req)
	if result == nil {
		var ext = filepath.Ext(req.URL.Path)
		if len(ext) > 0 {
			app.serveStaticFile(w, req, ext)
			return
		}
	}
	// handle error 404
	if result == nil {
		result = app.showError(req, 404)
	}
	res, ok := result.(*response)
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

func (app *application) AddErrorHandler(code int, handler Handler) {
	if app.errorHandlers == nil {
		app.errorHandlers = make(map[int]Handler)
	}
	app.errorHandlers[code] = handler
}

func (app *application) Route(strPth string, c interface{}) {
	if app.routeLocked {
		panic(errors.New("The controller cannot be added to app application after it is started."))
	}
	var t = reflect.TypeOf(c)
	cInfo := createControllerInfo(t)
	if app.router == nil {
		app.router = newRouter()
	}
	app.router.Handle(strPth, cInfo)
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
		appRoot = getCurrentDirectory()
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

	if !IsDir(webRoot) {
		return nil, errors.New("Path \"" + webRoot + "\" is not a directory")
	}
	app := &application{
		webRoot:     fixPath(webRoot),
		port:        port,
		initError:   nil,
		routeLocked: false,
	}
	err := app.init()
	return app, err
}
