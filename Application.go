package wemvc

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"os/exec"
	"path/filepath"
	"fmt"
)

type Application struct {
	webRoot string
	config  *configuration
}

func (this *Application) init() error {
	var configFile = this.MapPath("/web.config")
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	this.config = &configuration{}
	err = json.Unmarshal(bytes, this.config)
	if err != nil {
		return err
	}

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

func (this *Application) ServeHTTP(res http.ResponseWriter, req *http.Request) {

}

func (this *Application)Run() error {
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