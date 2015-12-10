package wemvc

import (
	"errors"
	"net/http"
)

// Context the request context interface
type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	RouteData() RouteData
}

type context struct {
	w          http.ResponseWriter
	req        *http.Request
	routeData  RouteData
	actionName string
	controller string
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
