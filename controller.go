package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"github.com/Simbory/wemvc/session"
)

// Controller the controller base struct
type Controller struct {
	Request     *http.Request
	Response    http.ResponseWriter
	RouteData   RouteData
	controller  string
	actionName  string
	session     session.Store
	sessionData map[string]interface{}
	ViewData    map[string]interface{}
	Items       map[string]interface{}
}

// OnInit the OnInit method is firstly called while executing the controller
func (ctrl *Controller) OnInit(req *http.Request, w http.ResponseWriter, controller, actionName string, routeData RouteData, ctxItems map[string]interface{}) {
	ctrl.Request = req
	ctrl.Response = w
	ctrl.RouteData = routeData
	if len(actionName) < 1 {
		ctrl.actionName = "index"
	} else {
		ctrl.actionName = actionName
	}
	ctrl.controller = controller
	ctrl.ViewData = make(map[string]interface{})
	if ctxItems != nil {
		ctrl.Items = ctxItems
	} else {
		ctrl.Items = make(map[string]interface{})
	}
}

// Session start the session
func (ctrl *Controller) Session() session.Store {
	if ctrl.session == nil {
		session,err := App.globalSession.SessionStart(ctrl.Response, ctrl.Request)
		if err != nil {
			panic(err)
		}
		ctrl.session = session
	}
	return ctrl.session
}

// OnLoad the OnLoad is called just after the OnInit
func (ctrl *Controller) OnLoad() {
}

// ViewFile execute a view file and return the HTML
func (ctrl *Controller) ViewFile(viewPath string) ActionResult {
	res, code := renderView(viewPath, ctrl.ViewData)
	var resp = NewActionResult()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

// View execute the default view file and return the HTML
func (ctrl *Controller) View() ActionResult {
	res, code := renderView(ctrl.controller+"/"+ctrl.actionName, ctrl.ViewData)
	var resp = NewActionResult()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

// Content return the content as text
func (ctrl *Controller) Content(str string, ctype ...string) ActionResult {
	var resp = NewActionResult()
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
func (ctrl *Controller) JSON(data interface{}) ActionResult {
	var resp = NewActionResult()
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("application/json")
	resp.Write(bytes)
	return resp
}

// XML return the Xml string as action result
func (ctrl *Controller) XML(obj interface{}) ActionResult {
	var resp = NewActionResult()
	bytes, err := xml.Marshal(obj)
	if err != nil {
		panic(err)
	}
	resp.SetContentType("text/xml")
	resp.Write(bytes)
	return resp
}

// File serve the file as action result
func (ctrl *Controller) File(path string, ctype string) ActionResult {
	var resp = &actionResult{
		statusCode:  200,
		resFile:     path,
		contentType: ctype,
	}
	return resp
}

// Redirect return a redirect url as action result
func (ctrl *Controller) Redirect(url string, statusCode ...int) ActionResult {
	var code = 302
	if len(statusCode) > 0 && statusCode[0] == 301 {
		code = 301
	}
	var resp = &actionResult{
		statusCode: code,
		redUrl:     url,
	}
	return resp
}

// NotFound return a 404 page as action result
func (ctrl *Controller) NotFound() ActionResult {
	return App.showError(ctrl.Request, 404)
}
