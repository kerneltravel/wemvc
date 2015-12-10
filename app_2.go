package wemvc

import (
	"errors"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/Simbory/wemvc/fsnotify"
)

// init app func is used to init the application
func (app *application) init() error {
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
	err = app.watcher.Watch(app.GetWebRoot())
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
	if IsDir(viewDir) {
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
	return nil
}

// watchFile is used to watching the required files: config files and view files
func (app *application) watchFile() {
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
				if IsDir(strFile) {
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

func (app *application) isConfigFile(f string) bool {
	if app.MapPath("/web.config") == f {
		return true
	}
	for _, configFile := range app.watchingFiles {
		if configFile == f {
			return true
		}
	}
	return false
}

func (app *application) isInViewFolder(f string) bool {
	var viewPath = app.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

func (app *application) loadConfig() (*configuration, []string, error) {
	// load the config file
	var configFile = app.MapPath("/web.config")
	if IsFile(configFile) == false {
		return nil, nil, nil
	}
	var configData = &configuration{}
	var files []string
	err := file2Xml(configFile, configData)
	if err != nil {
		return nil, nil, err
	}
	// load the setting config source file
	if len(configData.Settings.ConfigSource) > 0 {
		configFile = app.MapPath(configData.Settings.ConfigSource)
		var settings = &settingGroup{}
		err = file2Xml(configFile, settings)
		if err != nil {
			return nil, nil, err
		}
		configData.Settings.Settings = settings.Settings
		configData.Settings.ConfigSource = ""
		files = append(files, configFile)
	}
	// load the connection string config source
	if len(configData.ConnStrings.ConfigSource) > 0 {
		configFile = app.MapPath(configData.ConnStrings.ConfigSource)
		var conns = &connGroup{}
		err = file2Xml(configFile, conns)
		if err != nil {
			return nil, nil, err
		}
		configData.ConnStrings.ConnStrings = conns.ConnStrings
		configData.ConnStrings.ConfigSource = ""
		files = append(files, configFile)
	}
	// load the mime config source
	if len(configData.Mimes.ConfigSource) > 0 {
		configFile = app.MapPath(configData.Mimes.ConfigSource)
		var mimes = &mimeGroup{}
		err = file2Xml(configFile, mimes)
		if err != nil {
			return nil, nil, err
		}
		configData.Mimes.Mimes = mimes.Mimes
		configData.Mimes.ConfigSource = ""
		files = append(files, configFile)
	}
	return configData, files, nil
}

func (app *application) serveStaticFile(res http.ResponseWriter, req *http.Request, ext string) {
	http.ServeFile(res, req, app.MapPath(req.URL.Path))
}

func (app *application) serveDynamic(w http.ResponseWriter, req *http.Request) Response {
	var path = req.URL.Path
	var resp Response
	cInfo, routeData, match := app.router.Lookup(req.Method, path)
	if !match && cInfo != nil {
		var action = routeData.ByName("action")
		if len(action) < 1 {
			action = strings.ToLower(req.Method)
		} else {
			action = strings.ToLower(req.Method + action)
		}
		if cInfo.containsAction(action) {
			action = cInfo.actions[action]
			resp = app.execute(w, req, cInfo.controllerType, action, routeData)
		}
	}
	return resp
}

func (app *application) execute(w http.ResponseWriter, req *http.Request, t reflect.Type, action string, routeData RouteData) Response {
	var ctrl = reflect.New(t)
	var initMethod = ctrl.MethodByName("OnInit")
	cName := strings.ToLower(t.String())
	cName = strings.Split(cName, ".")[1]
	cName = strings.Replace(cName, "controller", "", -1)
	reg, _ := regexp.Compile("^" + strings.ToLower(req.Method))
	cAction := reg.ReplaceAllString(strings.ToLower(action), "")
	var ctx = &context{
		w:          w,
		req:        req,
		routeData:  routeData,
		actionName: cAction,
		controller: cName,
	}
	// call OnInit method
	initMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
	})
	//parse form
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if req.MultipartForm != nil {
			var size int64
			var maxSize = App.GetConfig().GetSetting("MaxFormSize")
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
	ctrl.MethodByName("OnLoad").Call(nil)
	// call action method
	m := ctrl.MethodByName(action)
	if !m.IsValid() {
		return nil
	}
	values := m.Call(nil)
	if len(values) == 1 {
		value, valid := values[0].Interface().(Response)
		if !valid {
			panic(errors.New("Invalid return type"))
		} else {
			return value
		}
	}
	return nil
}

func (app *application) error404(req *http.Request) Response {
	res := NewResponse()
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

func (app *application) error403(req *http.Request) Response {
	res := NewResponse()
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

func (app *application) showError(req *http.Request, code int) Response {
	var handler = app.errorHandlers[code]
	if handler != nil {
		return handler(req)
	}
	return app.error404(req)
}

func (app *application) viewFolder() string {
	return app.MapPath("/views")
}

func (app *application) panicRecover(res http.ResponseWriter, req *http.Request) {
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
