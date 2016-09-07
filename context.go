package wemvc

import (
	"net/http"
	"reflect"
)

// Context the request context interface
type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request

	CtrlType() reflect.Type
	Namespace() NamespaceSection
	ActionMethod() string
	ActionName() string
	CtrlName() string
	RouteData() map[string]string
	IsEnd() bool

	GetItem(key string) interface{}
	SetItem(key string, data interface{})
	EndContext()
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
	items        map[string]interface{}
	end          bool
	app          server
}

func (ctx *context) CtrlType() reflect.Type {
	return ctx.ctrlType
}

func (ctx *context) Namespace() NamespaceSection {
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
func (ctx *context) GetItem(key string) interface{} {
	if ctx.items == nil {
		return nil
	}
	return ctx.items[key]
}

// SetItem set the context item
func (ctx *context) SetItem(key string, data interface{}) {
	if ctx.items == nil {
		ctx.items = make(map[string]interface{})
	}
	ctx.items[key] = data
}

func (ctx *context) EndContext() {
	ctx.end = true
	panic(&endRequestError{})
}

func (ctx *context) IsEnd() bool {
	return ctx.end
}
