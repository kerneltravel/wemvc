package wemvc

import (
	"net/http"
	"encoding/json"
    "encoding/xml"
	"io/ioutil"
)

type IController interface {
	Init(Application, *http.Request,map[string]string)
	Get() Response
	Post() Response
	Delete() Response
	Head() Response
	Trace() Response
	Put() Response
	Options() Response
}

type Controller struct {
	App Application
	Request *http.Request
	RouteData map[string]string
	ViewData map[string]interface{}
}

func (this *Controller) Init(app Application, req *http.Request, routeData map[string]string) {
	this.App = app
	this.Request = req
	if routeData != nil {
		this.RouteData = routeData
	} else {
		this.RouteData = make(map[string]string)
	}
	this.ViewData = make(map[string]interface{})
}

func (this *Controller) Get() Response { return nil }

func (this *Controller) Post() Response { return nil }

func (this *Controller) Delete() Response { return nil }

func (this *Controller) Head() Response { return nil }

func (this *Controller) Trace() Response { return nil }

func (this *Controller) Put() Response { return nil }

func (this *Controller) Options() Response { return nil }

func (this *Controller) View(viewPath string) Response {
	res,err := renderView(viewPath, this.ViewData)
	var resp = NewResponse()
	if err == nil {
		resp.Write([]byte(res))
	} else {
		resp.Write([]byte(err.Error()))
		resp.SetStatusCode(500)
	}
	return resp
}

func (this *Controller) Content(str string, ctype ...string) Response {
	var resp = NewResponse()
	resp.Write([]byte(str))
	if len(ctype) > 0 && len(ctype[0]) > 0 {
		resp.SetContentType(ctype[0])
	} else {
		resp.SetContentType("text/plain")
	}
	return resp
}

func (this *Controller) Json(data interface{}) Response {
	var resp = NewResponse()
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("application/json")
	resp.Write(bytes)
	return resp
}

func (this *Controller) Xml(obj interface{}) Response {
	var resp = NewResponse()
	bytes, err := xml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("text/xml")
	resp.Write(bytes)
	return resp
}

func (this *Controller) File(path string, ctype string) Response {
	var resp = NewResponse()
	if isFile(path) {
		data,err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		} else {
			resp.Write(data)
			resp.SetContentType(ctype)
		}
	} else {
		resp.SetStatusCode(404)
	}
	return resp
}

func (this *Controller) Redirect(url string, statusCode ...int) Response {
	var resp = NewResponse()
	var code = 302
	if len(statusCode) > 0 && statusCode[0] == 301 {
		code = 301
	}
	resp.SetStatusCode(code)
	resp.SetHeader("Location", url)
	return resp
}

