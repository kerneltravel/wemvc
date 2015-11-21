package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type IController interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Context() Context
	RouteData() map[string]string
	OnInit(ctx *context)
	OnLoad()
}

type Controller struct {
	ctx      *context
	ViewData map[string]interface{}
}

func (this *Controller) Request() *http.Request {
	return this.ctx.Request()
}

func (this *Controller) Response() http.ResponseWriter {
	return this.ctx.Response()
}

func (this *Controller) RouteData() RouteData {
	return this.ctx.RouteData()
}

func (this *Controller) Context() Context {
	return this.ctx
}

func (this *Controller) OnInit(ctx *context) {
	this.ctx = ctx
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
	http.ServeFile(this.Response(), this.Request(), path)
	return this.Content("", ctype)
}

func (this *Controller) Redirect(url string, statusCode ...int) Response {
	var code = 302
	if len(statusCode) > 0 && statusCode[0] == 301 {
		code = 301
	}
	var red = &redirect{location: url, statusCode: code}
	panic(red)
	return NewResponse()
}

func (this *Controller) NotFound() Response {
	return App.(*application).showError(this.Request(), 404)
}
