package wemvc

import (
	"net/http"
	"reflect"
)

// Context the request context interface
type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	Namespace() NamespaceSection
	ActionMethod() string
	ActionName() string
	CtrlName() string
	RouteData() map[string]string
	CtxItems() *CtxItems
	Server() Server
	EndContext()
}

type CtxItems struct {
	items map[string]interface{}
}

func (ci *CtxItems) Get(key string) interface{} {
	return ci.items[key]
}

func (ci *CtxItems) Set(key string, data interface{}) {
	ci.items[key] = data
}

func (ci *CtxItems) Clear() {
	ci.items = nil
}

func (ci *CtxItems) Delete(key string) interface{} {
	data,ok := ci.items[key]
	if ok {
		delete(ci.items, key)
	}
	return data
}

type context struct {
	req          *http.Request
	w            http.ResponseWriter
	ctrlType     reflect.Type
	ns           string
	actionMethod string
	actionName   string
	ctrlName     string
	routeData    map[string]string
	ctxItems     *CtxItems
	app          *server
}

func (ctx *context) Server() Server {
	return ctx.app
}

func (ctx *context) Namespace() NamespaceSection {
	if len(ctx.ns) == 0 {
		return nil
	}
	return ctx.app.namespaces[ctx.ns]
}

func (ctx *context) ActionMethod() string {
	return ctx.actionMethod
}

func (ctx *context) ActionName() string {
	return ctx.actionName
}

func (ctx *context) CtrlName() string {
	return ctx.ctrlName
}

// Response get the response info
func (ctx *context) Response() http.ResponseWriter {
	if ctx.w == nil {
		panic(resEmptyError)
	}
	return ctx.w
}

/// Request get the request info
func (ctx *context) Request() *http.Request {
	if ctx.req == nil {
		panic(reqEmptyError)
	}
	return ctx.req
}

// RouteData get the route data
func (ctx *context) RouteData() map[string]string {
	if ctx.routeData == nil {
		ctx.routeData = make(map[string]string)
	}
	return ctx.routeData
}

// GetItem get the context item
func (ctx *context) CtxItems() *CtxItems {
	if ctx.ctxItems == nil {
		ctx.ctxItems = &CtxItems{items:make(map[string]interface{})}
	}
	return ctx.ctxItems
}

func (ctx *context) EndContext() {
	panic(&endRequestError{})
}
