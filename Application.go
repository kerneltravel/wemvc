package wemvc

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"io/ioutil"
	"time"
)

type Handler func(http.ResponseWriter, *http.Request)

type Application struct {
	ErrorHandlers map[int]Handler
	webRoot string
	config  *configuration
}

func (this *Application) init() error {
	// load the config file
	var configFile = this.MapPath("/web.config")
	this.config = &configuration{}
	err := file2Xml(configFile, this.config)
	if err != nil {
		return err
	}
	// load the setting config source file
	if len(this.config.Settings.ConfigSource) > 0 {
		configFile = this.MapPath(this.config.Settings.ConfigSource)
		var settings = &settingGroup{}
		err = file2Xml(configFile, settings)
		if err != nil {
			return err
		}
		this.config.Settings.Settings = settings.Settings
		this.config.Settings.ConfigSource = ""
	}
	// load the connection string config source
	if len(this.config.ConnStrings.ConfigSource) > 0 {
		configFile = this.MapPath(this.config.ConnStrings.ConfigSource)
		var conns = &connGroup{}
		err = file2Xml(configFile, conns)
		if err != nil {
			return err
		}
		this.config.ConnStrings.ConnStrings = conns.ConnStrings
		this.config.ConnStrings.ConfigSource = ""
	}
	// load the mime config source
	if len(this.config.Mimes.ConfigSource) > 0 {
		configFile = this.MapPath(this.config.Mimes.ConfigSource)
		var mimes = &mimeGroup{}
		err = file2Xml(configFile, mimes)
		this.config.Mimes.Mimes = mimes.Mimes
		this.config.Mimes.ConfigSource = ""
	}
	this.ErrorHandlers = make(map[int]Handler)
	this.ErrorHandlers[404] = this.error404
	return nil
}

func (this *Application) GetWebRoot() string {
	return this.webRoot
}

func (this *Application) GetConfig() Configuration {
	return this.config
}

func (this *Application) MapPath(relativePath string) string {
	var res = path.Join(this.GetWebRoot(), relativePath)
	return fixPath(res)
}

func (this *Application)serveFile(res http.ResponseWriter, req *http.Request, file string) {
	if !isFile(file) {
		this.showError(res, req, 404)
		return
	}
	var ext = filepath.Ext(file)
	if len(ext) < 1 {
		this.showError(res,req,404)
		return
	}
	var mime = this.GetConfig().GetMIME(ext)
	if len(mime) < 1 {
		this.showError(res,req,404)
		return
	}
	state,err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	var modifyTime = state.ModTime().Format(time.RFC1123)
	var ifMod = req.Header.Get("If-Modified-Since")
	if (modifyTime == ifMod) {
		res.WriteHeader(304)
		return
	}

	fbytes,err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res.Header().Add("Content-Type", mime)
	res.Header().Add("Last-Modified", modifyTime)
	res.WriteHeader(200)
	res.Write(fbytes)
}

func (this *Application)error404(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(404)
	res.Header().Add("Content-Type", "text/html;charset=utf-8")
	res.Write([]byte(`
	<div style="max-width:90%;margin:15px auto 0 auto;">
		<h1>ERROR 404</h1>
		<hr/>
		<p>The file you are looking for is not found!</p>
		<i>wemvc server version ` + version + `
	`))
}

func (this *Application)showError(res http.ResponseWriter, req *http.Request, code int) {
	var handler = this.ErrorHandlers[code]
	if handler != nil {
		handler(res, req)
	} else {
		this.error404(res,req)
	}
}

func (this *Application) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	defer func() {
		// Internal server handler
		if e := recover(); e != nil {
			res.WriteHeader(500)
			res.Header().Add("Content-Type", "text/html;charset=utf-8")
			res.Write([]byte(`
			<div style="max-width:90%;margin:15px auto 0 auto;">
				<h1>ERROR 500</h1>
				<hr/>
				<p>Internal server Error!</p>
				<p>` + e.(error).Error() + `</p>
				<i>wemvc server version ` + version + `
			</div>`))
		}
	}()

	var url = req.URL.Path
	var path = strings.TrimSuffix(this.MapPath(url), "/")
	
	var finalPath = ""
	if url == "" || isDir(path) {
		var defaultUrl = this.GetConfig().GetDefaultUrl()
		for _,f := range strings.Split(defaultUrl, ";") {
			path = path + "/" + f;
			if isFile(path) {
				finalPath = path
				break
			}
		}
	} else {
		finalPath = path
	}
	this.serveFile(res, req, finalPath)
}

func (this *Application) Run() error {
	port := fmt.Sprintf(":%d", this.config.Port)
	err := http.ListenAndServe(port, this)
	return err
}

func NewApplication(root string) (app *Application, err error) {
	if len(root) < 1 {
		err = errors.New("Web root cannot be empty.")
	}

	webRoot := strings.TrimSuffix(strings.TrimSuffix(root, "\\"), "/")
	if strings.HasPrefix(webRoot, ".") {
		file, _ := exec.LookPath(os.Args[0])
		exePath, _ := filepath.Abs(file)
		exeDir := filepath.Dir(exePath)
		webRoot = path.Join(exeDir, webRoot)
	}

	state, err := os.Stat(webRoot)
	if err != nil {
		return
	}
	if !state.IsDir() {
		err = errors.New("Path \"" + webRoot + "\" is not a directory")
	}
	app = &Application{webRoot: fixPath(webRoot)}
	err = app.init()
	if err != nil {
		app = nil
	}
	return
}
