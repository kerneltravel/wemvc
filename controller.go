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
	RouteData  map[string]string
	Controller string
	Action     string
	ViewData   map[string]interface{}
	Items      *CtxItems
	Namespace  *NsSection
	session    SessionStore
	Context    *Context
}

// OnInit this method is called at first while executing the controller
func (ctrl *Controller) OnInit(ctx *Context) {
	ctrl.Request = ctx.Request()
	ctrl.Response = ctx.Response()
	ctrl.RouteData = ctx.RouteData()
	ctrl.Namespace = ctx.Namespace()
	ctrl.Action = ctx.Ctrl.ActionName
	ctrl.Controller = ctx.Ctrl.ControllerName
	ctrl.ViewData = make(map[string]interface{})
	ctrl.Items = ctx.CtxItems()
	ctrl.Context = ctx
}

// Session start the session
func (ctrl *Controller) Session() SessionStore {
	if ctrl.session == nil {
		session, err := ctrl.Context.app.globalSession.SessionStart(ctrl.Response, ctrl.Request)
		if err != nil {
			panic(err)
		}
		ctrl.session = session
	}
	return ctrl.session
}

// ViewFile execute a view file and return the HTML
func (ctrl *Controller) ViewFile(viewPath string) Result {
	var res []byte
	ctrl.ViewData["Request"] = ctrl.Request
	ctrl.ViewData["Response"] = ctrl.Response
	ctrl.ViewData["RouteData"] = ctrl.RouteData
	ctrl.ViewData["Namespace"] = ctrl.Namespace
	ctrl.ViewData["Action"] = ctrl.Action
	ctrl.ViewData["Controller"] = ctrl.Controller
	ctrl.ViewData["CtxItems"] = ctrl.Items
	var err error
	if ctrl.Namespace != nil {
		res, err = ctrl.Namespace.renderView(viewPath, ctrl.ViewData)
	} else {
		res, err = ctrl.Context.app.renderView(viewPath, ctrl.ViewData)
	}
	if err != nil {
		panic(err)
	}
	var resp = NewResult()
	resp.Write(res)
	return resp
}

// View execute the default view file and return the HTML
func (ctrl *Controller) View() Result {
	return ctrl.ViewFile(ctrl.Controller + "/" + ctrl.Action)
}

// Content return the content as text
func (ctrl *Controller) Content(str string, cntType string) Result {
	var resp = NewResult()
	if len(cntType) < 1 {
		resp.ContentType = "text/plain"
	} else {
		resp.ContentType = cntType
	}
	if len(str) > 0 {
		resp.Write(str2Byte(str))
	}
	return resp
}

// PlainText return the text as plain text
func (ctrl *Controller) PlainText(content string) Result {
	return ctrl.Content(content, "text/plain")
}

// Javascript return the text as Javascript code
func (ctrl *Controller) Javascript(code string) Result {
	return ctrl.Content(code, "application/x-javascript")
}

// CSS return the text as css code
func (ctrl *Controller) CSS(code string) Result {
	return ctrl.Content(code, "text/css")
}

// JSON return the Json string as action result
func (ctrl *Controller) JSON(data interface{}) Result {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(byte2Str(bytes), "application/json")
}

// XML return the Xml string as action result
func (ctrl *Controller) XML(data interface{}) Result {
	bytes, err := xml.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(byte2Str(bytes), "text/xml")
}

// File serve the file as action result
func (ctrl *Controller) File(path string, cntType string) Result {
	var resp = &FileResult{
		FilePath:    path,
		ContentType: cntType,
	}
	return resp
}

func (ctrl *Controller) redirect(url string, statusCode int) *RedirectResult {
	var resp = &RedirectResult{
		StatusCode:  statusCode,
		RedirectUrl: url,
	}
	return resp
}

// Redirect Redirects a request to a new URL and specifies the new URL.
func (ctrl *Controller) Redirect(url string) Result {
	return ctrl.redirect(url, 302)
}

// RedirectPermanent Performs a permanent redirection from the requested URL to the specified URL.
func (ctrl *Controller) RedirectPermanent(url string) Result {
	return ctrl.redirect(url, 301)
}

// NotFound return a 404 page as action result
func (ctrl *Controller) NotFound() Result {
	return ctrl.Context.app.handleErrorReq(ctrl.Request, 404)
}

// EndRequest end the current request immediately
func (ctrl *Controller) EndRequest() {
	ctrl.Context.EndContext()
}
