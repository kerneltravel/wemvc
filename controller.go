package wemvc

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// Initializable indicate the controller can be initialized
type Initializable interface {
	OnInit(ctx *Context)
}

// Controller the controller base struct
type Controller struct {
	ViewData map[string]interface{}
	session  SessionStore
	ctx      *Context
}

// Request get the http request
func (ctrl *Controller) Request() *http.Request {
	return ctrl.ctx.Request()
}

// Response get the http response writer
func (ctrl *Controller) Response() http.ResponseWriter {
	return ctrl.ctx.Response()
}

// RouteData get the route data map
func (ctrl *Controller) RouteData() map[string]string {
	return ctrl.ctx.RouteData()
}

// Namespace get the current namespace
func (ctrl *Controller) Namespace() *NsSection {
	return ctrl.ctx.Namespace()
}

// ControllerName get the current controller name
func (ctrl *Controller) ControllerName() string {
	return ctrl.ctx.Ctrl.ControllerName
}

// ActionName get the current action name
func (ctrl *Controller) ActionName() string {
	return ctrl.ctx.Ctrl.ActionName
}

// Items get the current context items
func (ctrl *Controller) Items() *CtxItems {
	return ctrl.ctx.CtxItems()
}

// Cache get the current cache manager
func (ctrl *Controller) Cache() *CacheManager {
	return ctrl.ctx.app.cacheManager
}

//MapPath Returns the physical file path that corresponds to the specified virtual path.
func (ctrl *Controller) MapPath(virtualPath string) string {
	return ctrl.ctx.app.mapPath(virtualPath)
}

// Session start the session and get the session store
func (ctrl *Controller) Session() SessionStore {
	if ctrl.session == nil {
		session, err := ctrl.ctx.app.globalSession.SessionStart(ctrl.Response(), ctrl.Request())
		if err != nil {
			panic(err)
		}
		ctrl.session = session
	}
	return ctrl.session
}

// OnInit this method is called at first while executing the controller
func (ctrl *Controller) OnInit(ctx *Context) {
	ctrl.ViewData = make(map[string]interface{})
	ctrl.ctx = ctx
}

func (ctrl *Controller) initViewData() {
	ctrl.ViewData["Namespace"] = ctrl.Namespace()
	ctrl.ViewData["RouteData"] = ctrl.RouteData()
	ctrl.ViewData["Request"] = ctrl.Request()
	ctrl.ViewData["Session"] = ctrl.Session()
	ctrl.ViewData["Cache"] = ctrl.Cache()
}

// ViewFile execute a view file and return the HTML
func (ctrl *Controller) ViewFile(viewPath string) Result {
	var res []byte
	var err error
	ctrl.initViewData()
	if ctrl.Namespace() != nil {
		res, err = ctrl.Namespace().renderView(viewPath, ctrl.ViewData)
	} else {
		res, err = ctrl.ctx.app.renderView(viewPath, ctrl.ViewData)
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
	return ctrl.ViewFile(strAdd(ctrl.ControllerName(), "/", ctrl.ActionName()))
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
		RedirectURL: url,
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
	return ctrl.ctx.app.handleErrorReq(ctrl.Request(), 404)
}

// EndRequest end the current request immediately
func (ctrl *Controller) EndRequest() {
	ctrl.ctx.EndContext()
}
