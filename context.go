package wemvc

import (
	"errors"
	"net/http"
)

type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	RouteData() RouteData
}

type context struct {
	w         http.ResponseWriter
	req       *http.Request
	routeData RouteData
}

func (this *context) Response() http.ResponseWriter {
	if this.w == nil {
		panic(errors.New("response writer cannot be empty"))
	}
	return this.w
}

func (this *context) Request() *http.Request {
	if this.req == nil {
		panic(errors.New("http request cannot be empty"))
	}
	return this.req
}

func (this *context) RouteData() RouteData {
	if this.routeData == nil {
		this.routeData = RouteData{}
	}
	return this.routeData
}