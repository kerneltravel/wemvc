package wemvc

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sort"

	"github.com/Simbory/wemvc/fsnotify"
	"github.com/Simbory/wemvc/utils"
	"github.com/Simbory/wemvc/session"
	"fmt"
)


type server struct {
	errorHandlers map[int]Handler
	port          int
	webRoot       string
	config        *configuration
	router        *Router
	watcher       *fsnotify.Watcher
	watchingFiles []string
	initError     error
	routeLocked   bool
	staticPaths   []string
	filters       map[string][]Filter
	globalSession *session.SessionManager
}

func (app *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
		result = app.handleError(req, 404)
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

// init app func is used to init the application
func (app *server) init() error {
	// load the config file
	if config, f, err := app.loadConfig(); err != nil {
		app.initError = err
	} else {
		app.config = config
		app.watchingFiles = f
	}
	// init the error handler
	app.errorHandlers = make(map[int]Handler)
	app.errorHandlers[404] = app.error404
	app.errorHandlers[403] = app.error403
	// init fsnotify watcher
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	app.watcher = w
	// start to watch the config files
	err = app.watcher.Watch(RootDir())
	if err != nil {
		panic(err)
	}
	if app.initError == nil && len(app.watchingFiles) > 0 {
		for _, f := range app.watchingFiles {
			var dir = filepath.Dir(f)
			app.watcher.Watch(dir)
		}
	}
	// build the view template and watch the changes
	var viewDir = app.viewFolder()
	if utils.IsDir(viewDir) {
		buildViews(viewDir)
		app.watcher.Watch(viewDir)
		filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error {
			if info.IsDir() {
				app.watcher.Watch(p)
			}
			return nil
		})
	}
	// start to watch the files and dirs
	go app.watchFile()
	// init sessionManager
	if app.config.SessionConfig != nil && len(app.config.SessionConfig.ManagerName) > 0 {
		if app.config.SessionConfig.Gclifetime == 0 {
			app.config.SessionConfig.Gclifetime = 3600
		}
		if app.config.SessionConfig.Maxlifetime == 0 {
			app.config.SessionConfig.Maxlifetime = 3600
		}
		if app.config.SessionConfig.CookieLifeTime == 0 {
			app.config.SessionConfig.CookieLifeTime = 3600
		}
		mgr, err := session.NewManager(app.config.SessionConfig.ManagerName, app.config.SessionConfig)
		if err != nil {
			panic(err)
		}
		app.globalSession = mgr
		go app.globalSession.GC()
	}
	return nil
}

// watchFile is used to watching the required files: config files and view files
func (app *server) watchFile() {
	for {
		select {
		case ev := <-app.watcher.Event:
			strFile := path.Clean(ev.Name)
			if app.isConfigFile(strFile) {
				if config, f, err := app.loadConfig(); err != nil {
					app.initError = err
				} else {
					app.initError = nil
					app.config = config
					for _, configFile := range app.watchingFiles {
						app.watcher.RemoveWatch(configFile)
					}
					app.watchingFiles = f
					for _, f := range app.watchingFiles {
						app.watcher.Watch(f)
					}
				}
			} else if app.isInViewFolder(strFile) {
				if utils.IsDir(strFile) {
					if ev.IsDelete() {
						app.watcher.RemoveWatch(strFile)
					} else if ev.IsCreate() {
						app.watcher.Watch(strFile)
					}
				} else if strings.HasSuffix(strFile, ".html") {
					buildViews(app.viewFolder())
				}
			}
		}
	}
}

func (app *server) isConfigFile(f string) bool {
	if MapPath("/web.config") == f {
		return true
	}
	for _, configFile := range app.watchingFiles {
		if configFile == f {
			return true
		}
	}
	return false
}

func (app *server) isInViewFolder(f string) bool {
	var viewPath = app.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

func (app *server) loadConfig() (*configuration, []string, error) {
	// load the config file
	var configFile = MapPath("/web.config")
	if utils.IsFile(configFile) == false {
		return nil, nil, nil
	}
	var configData = &configuration{}
	var files []string
	err := utils.File2Xml(configFile, configData)
	if err != nil {
		return nil, nil, err
	}
	// load the setting config source file
	if len(configData.Settings.ConfigSource) > 0 {
		configFile = MapPath(configData.Settings.ConfigSource)
		var settings = &settingGroup{}
		err = utils.File2Xml(configFile, settings)
		if err != nil {
			return nil, nil, err
		}
		configData.Settings.Settings = settings.Settings
		configData.Settings.ConfigSource = ""
		files = append(files, configFile)
	}
	// load the connection string config source
	if len(configData.ConnStrings.ConfigSource) > 0 {
		configFile = MapPath(configData.ConnStrings.ConfigSource)
		var connGroup = &connGroup{}
		err = utils.File2Xml(configFile, connGroup)
		if err != nil {
			return nil, nil, err
		}
		configData.ConnStrings.ConnStrings = connGroup.ConnStrings
		configData.ConnStrings.ConfigSource = ""
		files = append(files, configFile)
	}
	// load the mime config source
	if len(configData.Mimes.ConfigSource) > 0 {
		configFile = MapPath(configData.Mimes.ConfigSource)
		var mimes = &mimeGroup{}
		err = utils.File2Xml(configFile, mimes)
		if err != nil {
			return nil, nil, err
		}
		configData.Mimes.Mimes = mimes.Mimes
		configData.Mimes.ConfigSource = ""
		files = append(files, configFile)
	}
	return configData, files, nil
}

func (app *server) serveStaticFile(res http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/") {
		var defaultUrls = Config().GetDefaultUrls()
		if len(defaultUrls) > 0 {
			for _, f := range defaultUrls {
				var file = MapPath(req.URL.Path + f)
				if utils.IsFile(file) {
					http.ServeFile(res, req, file)
					return
				}
			}
		} else {
			http.ServeFile(res, req, req.URL.Path+"index.html")
			return
		}
	}
	http.ServeFile(res, req, MapPath(req.URL.Path))
}

func (app *server) serveDynamic(ctx *context) ActionResult {
	var path = ctx.req.URL.Path
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimRight(path, "/")
	}
	var resp ActionResult
	cInfo, routeData, match := app.router.Lookup(ctx.req.Method, path)
	if !match && cInfo != nil {
		var action = routeData.ByName("action")
		var method = strings.ToTitle(ctx.req.Method)
		if len(action) < 1 {
			action = "index"
		}
		// find the action method in controller
		var actionMethod string
		if cInfo.containsAction(strings.ToLower(method + "_" + action)) {
			actionMethod = strings.ToLower(method + "_" + action)
		} else if cInfo.containsAction(strings.ToLower(method + action)) {
			actionMethod = strings.ToLower(method + action)
		} else if cInfo.containsAction(strings.ToLower(action)) {
			actionMethod = strings.ToLower(action)
		}
		if len(actionMethod) > 0 {
			actionMethod = cInfo.actions[actionMethod]
			ctx.routeData = routeData
			ctx.actionName = action
			// execute the action method
			resp = app.execute(ctx.req, ctx.w, cInfo.controllerType, actionMethod, action, routeData, ctx.items)
		}
	}
	return resp
}

func (app *server) execute(req *http.Request, w http.ResponseWriter, t reflect.Type, actionMethod, actionName string, routeData RouteData, items map[string]interface{}) ActionResult {
	var ctrl = reflect.New(t)
	cName := strings.ToLower(t.String())
	cName = strings.Split(cName, ".")[1]
	cName = strings.Replace(cName, "controller", "", -1)
	cAction := actionName

	// call OnInit method
	onInitMethod := ctrl.MethodByName("OnInit")
	if onInitMethod.IsValid() {
		onInitMethod.Call([]reflect.Value{
			reflect.ValueOf(req),
			reflect.ValueOf(w),
			reflect.ValueOf(cName),
			reflect.ValueOf(cAction),
			reflect.ValueOf(routeData),
			reflect.ValueOf(items),
		})
	}
	//parse form
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if req.MultipartForm != nil {
			var size int64
			var maxSize = app.config.GetSetting("MaxFormSize")
			if len(maxSize) < 1 {
				size = 10485760
			} else {
				size, _ = strconv.ParseInt(maxSize, 10, 64)
			}
			req.ParseMultipartForm(size)
		} else {
			req.ParseForm()
		}
	}
	// call OnLoad method
	onLoadMethod := ctrl.MethodByName("OnLoad")
	if onLoadMethod.IsValid() {
		onLoadMethod.Call(nil)
	}
	// call action method
	m := ctrl.MethodByName(actionMethod)
	if !m.IsValid() {
		return nil
	}
	values := m.Call(nil)
	if len(values) == 1 {
		value, valid := values[0].Interface().(ActionResult)
		if !valid {
			panic(errors.New("Invalid return type"))
		} else {
			return value
		}
	}
	return nil
}

func (app *server) error404(req *http.Request) ActionResult {
	res := NewActionResult()
	res.SetStatusCode(404)
	res.Write([]byte(`
	<div style="max-width:90%;margin:15px auto 0 auto;">
		<h1>ERROR 404</h1>
		<hr/>
		<p>The path "` + req.URL.Path + `" is not found!</p>
		<i>wemvc server version ` + Version + `</i>
	</div>`))
	return res
}

func (app *server) error403(req *http.Request) ActionResult {
	res := NewActionResult()
	res.SetStatusCode(403)
	res.Write([]byte(`
	<div style="max-width:90%;margin:15px auto 0 auto;">
		<h1>ERROR 403!</h1>
		<hr/>
		<p>Access denied for the path <b>` + req.URL.Path + `</b></p>
		<i>wemvc server version ` + Version + `</i>
	</div>`))
	return res
}

func (app *server) handleError(req *http.Request, code int) ActionResult {
	var handler = app.errorHandlers[code]
	if handler != nil {
		return handler(req)
	}
	return app.error404(req)
}

func (app *server) viewFolder() string {
	return MapPath("/views")
}

func (app *server) panicRecover(res http.ResponseWriter, req *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	res.WriteHeader(500)
	if err, ok := rec.(error); ok {
		res.Write([]byte(`
			<div style="max-width:90%;margin:15px auto 0 auto;">
				<h1>ERROR 500</h1>
				<hr/>
				<p>Internal server Error!</p>
				<p>` + err.Error() + `</p>
				<i>wemvc server version ` + Version + `</i>
			</div>`))
	} else {
		res.Write([]byte(`
			<div style="max-width:90%;margin:15px auto 0 auto;">
				<h1>ERROR 500</h1>
				<hr/>
				<p>Internal server Error!</p>
				<i>wemvc server version ` + Version + `</i>
			</div>`))
	}
}
