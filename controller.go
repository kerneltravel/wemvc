package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type IController interface {
	OnInit(ctx *context)
	OnLoad()
}

type Controller struct {
	Request    *http.Request
	Response   http.ResponseWriter
	RouteData  RouteData
	actionName string
	ViewData   map[string]interface{}
}

func (this *Controller) OnInit(ctx *context) {
	this.Request = ctx.req
	this.Response = ctx.w
	this.RouteData = ctx.routeData
	this.ViewData = make(map[string]interface{})
}

func (this *Controller) OnLoad() {
}

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
	var resp = &response{
		statusCode:  200,
		resFile: path,
		contentType: ctype,
	}
	return resp
}

func (this *Controller) Redirect(url string, statusCode ...int) Response {
	var code = 302
	if len(statusCode) > 0 && statusCode[0] == 301 {
		code = 301
	}
	var resp = &response{
		statusCode: code,
		redUrl: url,
	}
	return resp
}

func (this *Controller) NotFound() Response {
	return App.(*application).showError(this.Request, 404)
}