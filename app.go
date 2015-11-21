package wemvc

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Simbory/wemvc/fsnotify"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

type Handler func(*http.Request) Response

type Application interface {
	Port(...int) int
	AddErrorHandler(int, Handler)
	GetWebRoot() string
	GetConfig() Configuration
	MapPath(string) string
	Route(string, interface{}, ...string)
	Run() error
}

type application struct {
	errorHandlers map[int]Handler
	port          int
	webRoot       string
	config        *configuration
	route         routeTree
	watcher       *fsnotify.Watcher
	viewsWatcher  *fsnotify.Watcher
	watchingFiles []string
	initError     error
	routeLocked   bool
}

func (this *application) GetWebRoot() string {
	return this.webRoot
}

func (this *application) Port(p ...int) int {
	if len(p) > 0 {
		this.port = p[0]
	}
	return this.port
}

func (this *application) GetConfig() Configuration {
	return this.config
}

func (this *application) MapPath(relativePath string) string {
	var res = path.Join(this.GetWebRoot(), relativePath)
	return fixPath(res)
}

func (this *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// check init error
	if this.initError != nil {
		w.WriteHeader(500)
		w.Write([]byte(this.initError.Error()))
		return
	}
	defer this.panicRecover(w, req)

	// serve the dynamic page
	var result Response
	result = this.serveDynamic(w, req)
	if result == nil {
		var ext = filepath.Ext(req.URL.Path)
		if len(ext) > 0 {
			this.serveStaticFile(w, req, ext)
			return
		}
	}
	// handle error 404
	if result == nil {
		result = this.showError(req, 404)
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
	return
}

func (this *application) AddErrorHandler(code int, handler Handler) {
	if this.errorHandlers == nil {
		this.errorHandlers = make(map[int]Handler)
	}
	this.errorHandlers[code] = handler
}

func (this *application) Route(strPth string, c interface{}, v ...string) {
	if this.routeLocked {
		panic(errors.New("The controller cannot be added to this application after it is started."))
	}
	var t = reflect.TypeOf(c)
	cInfo := createControllerInfo(t)
	this.route.AddController(strPth, cInfo, v...)
}

func (this *application) Run() error {
	this.routeLocked = true
	port := fmt.Sprintf(":%d", this.port)
	err := http.ListenAndServe(port, this)
	return err
}

var App Application

func init() {
	port := flag.Int("port", 8080, "server running port")
	flag.Parse()
	app, err := newApp("wwwroot", *port)
	if err != nil {
		panic(err)
	}
	App = app
}

func newApp(root string, port int) (Application, error) {
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
