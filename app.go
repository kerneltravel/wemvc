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
	"strings"
)

type Handler func(*http.Request) Response

type Application interface {
	AddErrorHandler(int, Handler)
	GetWebRoot() string
	GetConfig() Configuration
	MapPath(string) string
	AddController(string, IController, ...string)
	Run() error
}

type application struct {
	errorHandlers map[int]Handler
	webRoot       string
	config        *configuration
	route         routeTree
}

func (this *application) GetWebRoot() string {
	return this.webRoot
}

func (this *application) GetConfig() Configuration {
	return this.config
}

func (this *application) MapPath(relativePath string) string {
	var res = path.Join(this.GetWebRoot(), relativePath)
	return fixPath(res)
}

func (this *application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//defer this.panicRecover(w)

	var result Response

	var ext = filepath.Ext(req.URL.Path)
	if len(ext) != 0 {
		// serve the static page
		result = this.serveStaticFile(req, ext)
	}
	// serve the dynamic page
	if result == nil {
		result = this.serveDynamic(req)
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
	var ctype = fmt.Sprintf("%s;charset=%s", result.GetContentType(), result.GetEncoding())
	w.Header().Add("Content-Type", ctype)
	if result.GetStatusCode() != 200 {
		w.WriteHeader(result.GetStatusCode())
	}
	w.Write(result.GetOutput())
	return
}

func (this *application) AddErrorHandler(code int, handler Handler) {
	if this.errorHandlers == nil {
		this.errorHandlers = make(map[int]Handler)
	}
	this.errorHandlers[code] = handler
}

func (this *application) AddController(strPth string, controller IController, v ...string) {
	this.route.AddController(strPth, controller, v...)
}

func (this *application) Run() error {
	port := fmt.Sprintf(":%d", this.config.Port)
	err := http.ListenAndServe(port, this)
	return err
}

var App Application

func init() {
	webroot := flag.String("root", "wwwroot", "the root path of the website")
	flag.Parse()
	app, err := newApp(*webroot)
	if err != nil {
		panic(err)
	}
	App = app
}

func newApp(root string) (Application, error) {
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

	if !isDir(webRoot) {
		return nil, errors.New("Path \"" + webRoot + "\" is not a directory")
	}
	app := &application{webRoot: fixPath(webRoot)}
	err := app.init()
	if err != nil {
		app = nil
	}
	return app, err
}
