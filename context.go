package wemvc

import (
	"net/http"
	"reflect"
)

type CtxRoute struct {
	RouteUrl  string
	RouteData map[string]string
	NsName    string
}

type CtxController struct {
	ControllerName string
	ControllerType reflect.Type

	ActionName       string
	ActionMethodName string
	ActionMethod     reflect.Value
}

type Context struct {
	req      *http.Request
	w        http.ResponseWriter
	ctxItems *CtxItems
	app      *server
	ended    bool

	Route  *CtxRoute
	Ctrl   *CtxController
	Result interface{}
}

/// Request get the request info
func (ctx *Context) Request() *http.Request {
	if ctx.req == nil {
		panic(errReqEmpty)
	}
	return ctx.req
}

// Response get the response info
func (ctx *Context) Response() http.ResponseWriter {
	if ctx.w == nil {
		panic(errRespEmpty)
	}
	return ctx.w
}

// GetItem get the context item
func (ctx *Context) CtxItems() *CtxItems {
	if ctx.ctxItems == nil {
		ctx.ctxItems = &CtxItems{items: make(map[string]interface{})}
	}
	return ctx.ctxItems
}

func (ctx *Context) Namespace() NsSection {
	if ctx.Route == nil || len(ctx.Route.NsName) == 0 {
		return nil
	}
	return ctx.app.namespaces[ctx.Route.NsName]
}

// RouteData get the route data
func (ctx *Context) RouteData() map[string]string {
	if ctx.Route == nil || ctx.Route.RouteData == nil {
		return nil
	}
	return ctx.Route.RouteData
}

func (ctx *Context) EndRequest() {
	panic(&errEndRequest{})
}

// EndContext end the request context now
func (ctx *Context) EndContext() {
	ctx.ended = true
}
