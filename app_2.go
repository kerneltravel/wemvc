package wemvc

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/Simbory/wemvc/fsnotify"
	"path"
	"path/filepath"
)

func (this *application) init() error {
	// load the config file
	if config,f,err := this.loadConfig();err != nil {
		println(err.Error())
		this.initError = err
	} else {
		this.config = config
		this.watchingFiles = f
	}
	// init the route tree
	this.route = routeTree{
		rootNode: routeNode{
			pathStr: "/",
			depth:   1,
		},
	}
	// build the view template
	buildViews(this.MapPath("/views"))
	// init the error handler
	this.errorHandlers = make(map[int]Handler)
	this.errorHandlers[404] = this.error404
	this.errorHandlers[403] = this.error403
	// init fsnotify watcher
	w,err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	this.watcher = w;

	err = this.watcher.Watch(this.MapPath("/webconfig.xml"))
	if err != nil {
		panic(err)
	}
	if this.initError == nil && len(this.watchingFiles) > 0 {
		for _, f := range this.watchingFiles {
			this.watcher.Watch(f)
		}
	}

	var viewDir = this.MapPath("/views")
	this.watcher.Watch(viewDir)
	filepath.Walk(viewDir, func(p string, info os.FileInfo, er error) error{
		println(p)
		if info.IsDir() {
			this.watcher.Watch(p)
		}
		return nil
	})
	go this.watchConfig()
	return nil
}

func (this *application)watchConfig() {
	for {
		select {
		case ev := <-this.watcher.Event:
			println(ev.Name)
			strFile := path.Clean(ev.Name)
			if this.isConfigFile(strFile) {
				if config,f,err := this.loadConfig();err != nil {
					this.initError = err
					println(err.Error())
				} else {
					this.initError = nil
					this.config = config
					for _, configFile := range this.watchingFiles {
						this.watcher.RemoveWatch(configFile)
					}
					this.watchingFiles = f
					for _, f := range this.watchingFiles {
						println(f)
						this.watcher.Watch(f)
					}
				}
			} else if this.isViewFile(strFile) {
				println(ev.String())
				if ev.IsDelete() {
					this.watcher.RemoveWatch(ev.Name)
				} else if ev.IsCreate() {
					this.watcher.AddWatch(ev.Name, fsnotify.FSN_ALL)
				}
				buildViews(this.MapPath("/views"))
			}
		}
	}
}

func (this *application)isConfigFile(f string) bool {
	if this.MapPath("/webconfig.xml") == f {
		return true
	}
	for _, configFile := range this.watchingFiles {
		if configFile == f {
			return true
		}
	}
	return false
}

func (this *application)isViewFile(f string) bool {
	var viewPath = this.MapPath("/views")
	return strings.HasPrefix(f, viewPath)
}

func (this *application)loadConfig() (*configuration, []string, error) {
	// load the config file
	var configFile = this.MapPath("/webconfig.xml")
	var configData = &configuration{}
	var files []string
	err := file2Xml(configFile, configData)
	if err != nil {
		return nil,nil,err
	}
	// load the setting config source file
	if len(configData.Settings.ConfigSource) > 0 {
		configFile = this.MapPath(configData.Settings.ConfigSource)
		var settings = &settingGroup{}
		err = file2Xml(configFile, settings)
		if err != nil {
			return nil,nil,err
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
			return nil,nil,err
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
			return nil,nil,err
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
			return nil,nil,err
		}
		configData.ProtectionUrls.ProtectionUrls = protectGroup.ProtectionUrls
		configData.ProtectionUrls.ConfigSource = ""
		files = append(files, configFile)
	}
	return configData,files,nil
}

func (this *application) serveStaticFile(req *http.Request, ext string) Response {
	var mime = this.GetConfig().GetMIME(ext)
	if len(mime) < 1 {
		return nil
	}
	var url = req.URL.Path
	if this.urlProtected(url) {
		return this.showError(req, 403)
	}
	var path = strings.TrimSuffix(this.MapPath(url), "/")
	var file = ""
	if url == "" || isDir(path) {
		var defaultUrl = this.GetConfig().GetDefaultUrl()
		for _, f := range strings.Split(defaultUrl, ";") {
			path = path + "/" + f
			if isFile(path) {
				file = path
				break
			}
		}
	} else {
		file = path
	}
	if !isFile(file) {
		return nil
	}
	state, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	var res = NewResponse()
	var modifyTime = state.ModTime().Format(time.RFC1123)
	var ifMod = req.Header.Get("If-Modified-Since")
	if modifyTime == ifMod {
		res.SetStatusCode(304)
		return res
	}
	fbytes, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res.SetContentType(mime)
	res.SetHeader("Last-Modified", modifyTime)
	res.Write(fbytes)
	return res
}

func (this *application) serveDynamic(req *http.Request) Response {
	var path = req.URL.Path
	var pathUrls []string
	if path == "/" {
		pathUrls = []string{"/"}
	} else {
		pathUrls = strings.Split(path, "/")
		pathUrls[0] = "/"
	}
	var resp Response = nil
	var routeData = make(map[string]string)
	res, c := this.route.rootNode.matchDepth(pathUrls, routeData)
	if res && c != nil {
		c.Init(this, req, routeData)
		if GET.Equal(req.Method) {
			resp = c.Get()
		} else if POST.Equal(req.Method) {
			resp = c.Post()
		} else if DELETE.Equal(req.Method) {
			resp = c.Delete()
		} else if HEAD.Equal(req.Method) {
			resp = c.Head()
		} else if TRACE.Equal(req.Method) {
			resp = c.Trace()
		} else if PUT.Equal(req.Method) {
			resp = c.Put()
		} else if OPTIONS.Equal(req.Method) {
			resp = c.Options()
		}
	}
	//panic(errors.New("test"))
	//panic(&redirect{location: "http://www.baidu.com/index.html"})
	return resp
}

func (this *application) error404(req *http.Request) Response {
	res := NewResponse()
	res.SetStatusCode(404)
	res.Write([]byte(`
	<div style="max-width:90%;margin:15px auto 0 auto;">
		<h1>ERROR 404</h1>
		<hr/>
		<p>The file you are looking for is not found!</p>
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
		} else if isDir(this.MapPath(url)) && strings.HasPrefix(url+"/", s.GetPathPrefix()) {
			return true
		}
	}
	return false
}

func (this *application) panicRecover(res http.ResponseWriter) {
	rec := recover()
	if rec == nil {
		return
	}

	if re, ok := rec.(*redirect); ok {
		res.Header().Set("Location", re.location)
		res.WriteHeader(302)
	} else {
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
		res.WriteHeader(500)
	}
}
