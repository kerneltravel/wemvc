package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Controller the controller base struct
type Controller struct {
	Request    *http.Request
	Response   http.ResponseWriter
	RouteData  RouteData
	controller string
	actionName string
	ViewData   map[string]interface{}
}

// OnInit the OnInit method is firstly called while executing the controller
func (ctrl *Controller) OnInit(ctx *context) {
	ctrl.Request = ctx.req
	ctrl.Response = ctx.w
	ctrl.RouteData = ctx.routeData
	if len(ctx.actionName) < 1 {
		ctrl.actionName = "index"
	} else {
		ctrl.actionName = ctx.actionName
	}
	ctrl.controller = ctx.controller
	ctrl.ViewData = make(map[string]interface{})
}

// OnLoad the OnLoad is called just after the OnInit
func (ctrl *Controller) OnLoad() {
}

// ViewFile execute a view file and return the HTML
func (ctrl *Controller) ViewFile(viewPath string) Response {
	res, code := renderView(viewPath, ctrl.ViewData)
	var resp = NewResponse()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

// View execute the default view file and renturn the HTML
func (ctrl *Controller) View() Response {
	res, code := renderView(ctrl.controller+"/"+ctrl.actionName, ctrl.ViewData)
	var resp = NewResponse()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

// Content return the content as text
func (ctrl *Controller) Content(str string, ctype ...string) Response {
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

// JSON return the Json string as action result
func (ctrl *Controller) JSON(data interface{}) Response {
	var resp = NewResponse()
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("application/json")
	resp.Write(bytes)
	return resp
}

// XML return the Xml string as action result
func (ctrl *Controller) XML(obj interface{}) Response {
	var resp = NewResponse()
	bytes, err := xml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("text/xml")
	resp.Write(bytes)
	return resp
}

// File serve the file as action result
func (ctrl *Controller) File(path string, ctype string) Response {
	var resp = &response{
		statusCode:  200,
		resFile:     path,
		contentType: ctype,
	}
	return resp
}

// Redirect return a redirect url as action result
func (ctrl *Controller) Redirect(url string, statusCode ...int) Response {
	var code = 302
	if len(statusCode) > 0 && statusCode[0] == 301 {
		code = 301
	}
	var resp = &response{
		statusCode: code,
		redUrl:     url,
	}
	return resp
}

// NotFound return a 404 page as action result
func (ctrl *Controller) NotFound() Response {
	return App.showError(ctrl.Request, 404)
}
