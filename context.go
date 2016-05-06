package wemvc

import (
	"errors"
	"net/http"
)

// Context the request context interface
type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	GetItem(key string) interface{}
	SetItem(key string, data interface{})
	End()
}

type context struct {
	end        bool
	w          http.ResponseWriter
	req        *http.Request
	routeData  RouteData
	actionName string
	controller string
	items      map[string]interface{}
}

func (ctx *context) Response() http.ResponseWriter {
	if ctx.w == nil {
		panic(errors.New("response writer cannot be empty"))
	}
	return ctx.w
}

func (ctx *context) Request() *http.Request {
	if ctx.req == nil {
		panic(errors.New("http request cannot be empty"))
	}
	return ctx.req
}

func (ctx *context) RouteData() RouteData {
	if ctx.routeData == nil {
		ctx.routeData = RouteData{}
	}
	return ctx.routeData
}

func (ctx *context) GetItem(key string) interface{} {
	if ctx.items == nil {
		return nil
	}
	return ctx.items[key]
}

func (ctx *context) SetItem(key string, data interface{}) {
	if ctx.items == nil {
		ctx.items = make(map[string]interface{})
	}
	ctx.items[key] = data
}

func (ctx *context) End() {
	ctx.end = true
}
