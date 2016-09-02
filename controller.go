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
	RouteData  RouteData
	Controller string
	Action     string
	ViewData   map[string]interface{}
	Items      map[string]interface{}
	app        server
	ns         string
	session    SessionStore
}

// OnInit this method is called at first while executing the controller
func (ctrl *Controller) OnInit(app server, req *http.Request, w http.ResponseWriter, ns, controller, actionName string, routeData RouteData, ctxItems map[string]interface{}) {
	ctrl.app = app
	ctrl.Request = req
	ctrl.Response = w
	ctrl.RouteData = routeData
	ctrl.ns = ns
	ctrl.Action = actionName
	ctrl.Controller = controller
	ctrl.ViewData = make(map[string]interface{})
	if ctxItems != nil {
		ctrl.Items = ctxItems
	} else {
		ctrl.Items = make(map[string]interface{})
	}
}

// Session start the session
func (ctrl *Controller) Session() SessionStore {
	if ctrl.session == nil {
		session, err := ctrl.app.globalSession.SessionStart(ctrl.Response, ctrl.Request)
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
	var res template.HTML
	var code int
	if len(ctrl.ns) > 0 {
		ns := ctrl.app.namespaces[ctrl.ns]
		if ns != nil {
			ctrl.initViewData()
			res, code = ns.renderView(viewPath, ctrl.ViewData)
		} else {
			return ctrl.NotFound()
		}
	} else {
		ctrl.initViewData()
		res, code = ctrl.app.renderView(viewPath, ctrl.ViewData)
	}
	var resp = NewActionResult()
	resp.Write([]byte(res))
	if code != 200 {
		resp.SetStatusCode(code)
	}
	return resp
}

func (ctrl *Controller) initViewData() {
	ctrl.ViewData["Request"] = ctrl.Request
}

// Namespace return the namespace using in this controller
func (ctrl *Controller) Namespace() NamespaceSection {
	if len(ctrl.ns) < 1 {
		return nil
	}
	return ctrl.app.namespaces[ctrl.ns]
}

// View execute the default view file and return the HTML
func (ctrl *Controller) View() ActionResult {
	return ctrl.ViewFile(ctrl.Controller + "/" + ctrl.Action)
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

// Javascript return the text as Javascript code
func (ctrl *Controller) Javascript(code string) ActionResult {
	return ctrl.Content(code, "application/x-javascript")
}

// CSS return the text as css code
func (ctrl *Controller) CSS(code string) ActionResult {
	return ctrl.Content(code, "text/css")
}

// JSON return the Json string as action result
func (ctrl *Controller) JSON(data interface{}) ActionResult {
	bytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return ctrl.Content(string(bytes), "application/json")
}

// XML return the Xml string as action result
func (ctrl *Controller) XML(data interface{}) ActionResult {
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
	return ctrl.app.handleError(ctrl.Request, 404)
}

func (ctrl *Controller) EndRequest() {
	panic(&endRequestError{})
}
