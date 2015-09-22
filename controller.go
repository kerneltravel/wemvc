package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type IController interface {
	Init(http.ResponseWriter, *http.Request, map[string]string)
	Get() Response
	Post() Response
	Delete() Response
	Head() Response
	Trace() Response
	Put() Response
	Options() Response
}

type Controller struct {
	Request   *http.Request
	Response  http.ResponseWriter
	RouteData map[string]string
	ViewData  map[string]interface{}
}

func (this *Controller) Init(w http.ResponseWriter, req *http.Request, routeData map[string]string) {
	this.Request = req
	this.Response = w
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
	res, code := renderView(viewPath, this.ViewData)
	var resp = NewResponse()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

func (this *Controller) Content(str string, ctype ...string) Response {
	var resp = NewResponse()
	if len(str) > 0 {
		resp.Write([]byte(str))
	}
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
	http.ServeFile(this.Response, this.Request, path)
	return this.Content("", ctype)
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

func (this *Controller) NotFound() Response {
	return App.(*application).showError(this.Request, 404)
}
