package wemvc

import (
	"errors"
	"github.com/Simbory/wemvc/fsnotify"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

func (this *application) init() error {
	// load the config file
	if config, f, err := this.loadConfig(); err != nil {
		this.initError = err
	} else {
		this.config = config
		this.watchingFiles = f
	}
	// build the view template
	buildViews(this.viewFolder())
	// init the error handler
	this.errorHandlers = make(map[int]Handler)
	this.errorHandlers[404] = this.error404
	this.errorHandlers[403] = this.error403
	// init fsnotify watcher
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	this.watcher = w

	err = this.watcher.Watch(this.GetWebRoot())
	if err != nil {
		panic(err)
	}
	if this.initError == nil && len(this.watchingFiles) > 0 {
		for _, f := range this.watchingFiles {
			var dir = filepath.Dir(f)
			this.watcher.Watch(dir)
		}
	}

	var viewDir = this.viewFolder()
	this.watcher.Watch(viewDir)
	filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error {
		if info.IsDir() {
			this.watcher.Watch(p)
		}
		return nil
	})
	go this.watchFile()
	return nil
}

func (this *application) watchFile() {
	for {
		select {
		case ev := <-this.watcher.Event:
			strFile := path.Clean(ev.Name)
			if this.isConfigFile(strFile) {
				if config, f, err := this.loadConfig(); err != nil {
					this.initError = err
				} else {
					this.initError = nil
					this.config = config
					for _, configFile := range this.watchingFiles {
						this.watcher.RemoveWatch(configFile)
					}
					this.watchingFiles = f
					for _, f := range this.watchingFiles {
						this.watcher.Watch(f)
					}
				}
			} else if this.isInViewFolder(strFile) {
				if IsDir(strFile) {
					if ev.IsDelete() {
						this.watcher.RemoveWatch(strFile)
					} else if ev.IsCreate() {
						this.watcher.Watch(strFile)
					}
				} else if strings.HasSuffix(strFile, ".html") {
					buildViews(this.viewFolder())
				}
			}
		}
	}
}

func (this *application) isConfigFile(f string) bool {
	if this.MapPath("/web.config") == f {
		return true
	}
	for _, configFile := range this.watchingFiles {
		if configFile == f {
			return true
		}
	}
	return false
}

func (this *application) isInViewFolder(f string) bool {
	var viewPath = this.viewFolder()
	return strings.HasPrefix(f, viewPath)
}

func (this *application) loadConfig() (*configuration, []string, error) {
	// load the config file
	var configFile = this.MapPath("/web.config")
	var configData = &configuration{}
	var files []string
	err := file2Xml(configFile, configData)
	if err != nil {
		return nil, nil, err
	}
	// load the setting config source file
	if len(configData.Settings.ConfigSource) > 0 {
		configFile = this.MapPath(configData.Settings.ConfigSource)
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
		configFile = this.MapPath(configData.ConnStrings.ConfigSource)
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
		configFile = this.MapPath(configData.Mimes.ConfigSource)
		var mimes = &mimeGroup{}
		err = file2Xml(configFile, mimes)
		if err != nil {
			return nil, nil, err
		}
		configData.Mimes.Mimes = mimes.Mimes
		configData.Mimes.ConfigSource = ""
		files = append(files, configFile)
	}
	// load the protection url setting
	if len(configData.ProtectionUrls.ConfigSource) > 0 {
		configFile = this.MapPath(configData.ProtectionUrls.ConfigSource)
		var protectGroup = &protectionUrlGroup{}
		err = file2Xml(configFile, protectGroup)
		if err != nil {
			return nil, nil, err
		}
		configData.ProtectionUrls.ProtectionUrls = protectGroup.ProtectionUrls
		configData.ProtectionUrls.ConfigSource = ""
		files = append(files, configFile)
	}
	return configData, files, nil
}

func (this *application) serveStaticFile(res http.ResponseWriter, req *http.Request, ext string) {
	var mime = this.GetConfig().GetMIME(ext)
	if len(mime) < 1 {
		var r = this.showError(req, 404)
		if r == nil {
			r = this.error404(req)
		}
		res.WriteHeader(r.GetStatusCode())
		res.Write(r.GetOutput())
		return
	}
	var url = req.URL.Path
	if this.urlProtected(url) {
		var r = this.showError(req, 403)
		if r == nil {
			r = this.error403(req)
		}
		res.WriteHeader(r.GetStatusCode())
		res.Write(r.GetOutput())
		return
	}
	http.ServeFile(res, req, this.MapPath(req.URL.Path))
}

func (this *application) serveDynamic(w http.ResponseWriter, req *http.Request) Response {
	var path = req.URL.Path
	var resp Response = nil
	cInfo, routeData, match := this.router.Lookup(req.Method, path)
	if !match {
		var action = routeData.ByName("action")
		if len(action) < 1 {
			action = strings.ToLower(req.Method)
		} else {
			action = strings.ToLower(req.Method + action)
		}
		if cInfo.containsAction(action) {
			action = cInfo.actions[action]
			resp = this.execute(w, req, cInfo.controllerType, action, routeData)
		}
	}
	return resp
}

func (this *application) execute(w http.ResponseWriter, req *http.Request, t reflect.Type, action string, routeData RouteData) Response {
	var ctrl = reflect.New(t)
	var initMethod = ctrl.MethodByName("OnInit")
	var ctx = &context{
		w:         w,
		req:       req,
		routeData: routeData,
	}
	// call OnInit method
	initMethod.Call([]reflect.Value{
		reflect.ValueOf(ctx),
	})
	//parse form
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		if req.MultipartForm != nil {
			var size int64 = 0
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

func (this *application) error404(req *http.Request) Response {
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

func (this *application) error403(req *http.Request) Response {
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

func (this *application) showError(req *http.Request, code int) Response {
	var handler = this.errorHandlers[code]
	if handler != nil {
		return handler(req)
	} else {
		return this.error404(req)
	}
}

func (this *application) urlProtected(url string) bool {
	for _, s := range this.GetConfig().GetProtectionUrls() {
		if strings.HasPrefix(url, s.GetPathPrefix()) {
			return true
		} else if IsDir(this.MapPath(url)) && strings.HasPrefix(url+"/", s.GetPathPrefix()) {
			return true
		}
	}
	return false
}

func (this *application) viewFolder() string {
	return this.MapPath("/views")
}

func (this *application) panicRecover(res http.ResponseWriter, req *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	if re, ok := rec.(*redirect); ok {
		http.Redirect(res, req, re.location, re.statusCode)
	} else {
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
}
