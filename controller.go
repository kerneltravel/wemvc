package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
)

// Controller the controller base struct
type Controller struct {
	Request    *http.Request
	Response   http.ResponseWriter
	RouteData  map[string]string
	Controller string
	Action     string
	ViewData   map[string]interface{}
	Items      *CtxItems
	Server     Server
	Namespace  NamespaceSection
	session    SessionStore
	Context    Context
}

// OnInit this method is called at first while executing the controller
func (ctrl *Controller) OnInit(ctx Context) {
	ctrl.Server = ctx.Server()
	ctrl.Request = ctx.Request()
	ctrl.Response = ctx.Response()
	ctrl.RouteData = ctx.RouteData()
	ctrl.Namespace = ctx.Namespace()
	ctrl.Action = ctx.ActionName()
	ctrl.Controller = ctx.CtrlName()
	ctrl.ViewData = make(map[string]interface{})
	ctrl.Items = ctx.CtxItems()
	ctrl.Context = ctx
}

// Session start the session
func (ctrl *Controller) Session() SessionStore {
	if ctrl.session == nil {
		session, err := ctrl.Server.(*server).globalSession.SessionStart(ctrl.Response, ctrl.Request)
		if err != nil {
			panic(err)
		}
		ctrl.session = session
	}
	return ctrl.session
}

// OnLoad the OnLoad is called just after the OnInit method
func (ctrl *Controller) OnLoad() {
}

// ViewFile execute a view file and return the HTML
func (ctrl *Controller) ViewFile(viewPath string) *Result {
	var res template.HTML
	var code int
	ctrl.ViewData["Server"] = ctrl.Server
	ctrl.ViewData["Request"] = ctrl.Request
	ctrl.ViewData["Response"] = ctrl.Response
	ctrl.ViewData["RouteData"] = ctrl.RouteData
	ctrl.ViewData["Namespace"] = ctrl.Namespace
	ctrl.ViewData["Action"] = ctrl.Action
	ctrl.ViewData["Controller"] = ctrl.Controller
	ctrl.ViewData["CtxItems"] = ctrl.Items
	if ctrl.Namespace != nil {
		res, code = ctrl.Namespace.(*namespace).renderView(viewPath, ctrl.ViewData)
	} else {
		res, code = ctrl.Server.(*server).renderView(viewPath, ctrl.ViewData)
	}
	var resp = NewResult()
	resp.Write([]byte(res))
	if code != 200 {
		resp.StatusCode = code
	}
	return resp
}

// View execute the default view file and return the HTML
func (ctrl *Controller) View() *Result {
	return ctrl.ViewFile(ctrl.Controller + "/" + ctrl.Action)
}

// Content return the content as text
func (ctrl *Controller) Content(str string, cntType string) *Result {
	var resp = NewResult()
	if len(cntType) < 1 {
		resp.ContentType = "text/plain"
	} else {
		resp.ContentType = cntType
	}
	if len(str) > 0 {
		resp.Write([]byte(str))
	}
	return resp
}

// PlainText return the text as plain text
func (ctrl *Controller) PlainText(content string) *Result {
	return ctrl.Content(content, "text/plain")
}

// Javascript return the text as Javascript code
func (ctrl *Controller) Javascript(code string) *Result {
	return ctrl.Content(code, "application/x-javascript")
}

// CSS return the text as css code
func (ctrl *Controller) CSS(code string) *Result {
	return ctrl.Content(code, "text/css")
}

// JSON return the Json string as action result
func (ctrl *Controller) JSON(data interface{}) *Result {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(string(bytes), "application/json")
}

// XML return the Xml string as action result
func (ctrl *Controller) XML(data interface{}) *Result {
	bytes, err := xml.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(string(bytes), "text/xml")
}

// File serve the file as action result
func (ctrl *Controller) File(path string, cntType string) *Result {
	var resp = &Result{
		StatusCode:  200,
		respFile:     path,
		ContentType: cntType,
	}
	return resp
}

func (ctrl *Controller) redirect(url string, statusCode int) *Result {
	var resp = &Result{
		StatusCode: statusCode,
		redURL:     url,
	}
	return resp
}

// Redirect Redirects a request to a new URL and specifies the new URL.
func (ctrl *Controller) Redirect(url string) *Result {
	return ctrl.redirect(url, 302)
}

// RedirectPermanent Performs a permanent redirection from the requested URL to the specified URL.
func (ctrl *Controller) RedirectPermanent(url string) *Result {
	return ctrl.redirect(url, 301)
}

// NotFound return a 404 page as action result
func (ctrl *Controller) NotFound() *Result {
	return ctrl.Server.(*server).handleError(ctrl.Request, 404)
}

// EndRequest end the current request immediately
func (ctrl *Controller) EndRequest() {
	ctrl.Context.EndContext()
}
