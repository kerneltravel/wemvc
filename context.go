package wemvc

import (
	"errors"
	"net/http"
)

type Context interface {
	Response() http.ResponseWriter
	Request() *http.Request
	RouteData() map[string]string
}

type context struct {
	w         http.ResponseWriter
	req       *http.Request
	routeData map[string]string
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

func (this *context) RouteData() map[string]string {
	if this.routeData == nil {
		this.routeData = make(map[string]string)
	}
	return this.routeData
}
