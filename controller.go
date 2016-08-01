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

// OnInit this method is called at first while executing the controller
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
		session, err := app.globalSession.SessionStart(ctrl.Response, ctrl.Request)
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
func (ctrl *Controller) ViewFile(viewPath string) ActionResult {
	res, code := app.renderView(viewPath, ctrl.ViewData)
	var resp = NewActionResult()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

// View execute the default view file and return the HTML
func (ctrl *Controller) View() ActionResult {
	return ctrl.ViewFile(ctrl.controller + "/" + ctrl.actionName)
}

// Content return the content as text
func (ctrl *Controller) Content(str string, cntType string) ActionResult {
	var resp = NewActionResult()
	if len(cntType) < 1 {
		resp.SetContentType("text/plain")
	} else {
		resp.SetContentType(cntType)
	}
	if len(str) > 0 {
		resp.Write([]byte(str))
	}
	return resp
}

// PlainText return the text as plain text
func (ctrl *Controller) PlainText(content string) ActionResult {
	return ctrl.Content(content, "text/plain")
}

func (ctrl *Controller) Javascript(code string) ActionResult {
	return ctrl.Content(code, "application/x-javascript")
}

func (ctrl *Controller) Css(code string) ActionResult {
	return ctrl.Content(code, "text/css")
}

// Json return the Json string as action result
func (ctrl *Controller) Json(data interface{}) ActionResult {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(string(bytes), "application/json")
}

// Xml return the Xml string as action result
func (ctrl *Controller) Xml(data interface{}) ActionResult {
	bytes, err := xml.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(string(bytes), "text/xml")
}

// File serve the file as action result
func (ctrl *Controller) File(path string, cntType string) ActionResult {
	var resp = &actionResult{
		statusCode:  200,
		resFile:     path,
		contentType: cntType,
	}
	return resp
}

func (ctrl *Controller) redirect(url string, statusCode int) ActionResult {
	var resp = &actionResult{
		statusCode: statusCode,
		redURL:     url,
	}
	return resp
}

// Redirect Redirects a request to a new URL and specifies the new URL.
func (ctrl *Controller) Redirect(url string) ActionResult {
	return ctrl.redirect(url, 302)
}

// RedirectPermanent Performs a permanent redirection from the requested URL to the specified URL.
func (ctrl *Controller) RedirectPermanent(url string) ActionResult {
	return ctrl.redirect(url, 301)
}

// NotFound return a 404 page as action result
func (ctrl *Controller) NotFound() ActionResult {
	return app.handleError(ctrl.Request, 404)
}
