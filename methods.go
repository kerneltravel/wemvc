package wemvc

import "strings"

type HttpMethod string

func (this HttpMethod) Equal(method string) bool {
	return strings.ToLower(string(this)) == strings.ToLower(method)
}

func (this HttpMethod) String() string {
	return strings.ToUpper(string(this))
}

const (
	GET     = HttpMethod("GET")
	POST    = HttpMethod("POST")
	DELETE  = HttpMethod("DELETE")
	HEAD    = HttpMethod("HEAD")
	TRACE   = HttpMethod("TRACE")
	PUT     = HttpMethod("PUT")
	OPTIONS = HttpMethod("OPTIONS")
	CONNECT = HttpMethod("CONNECT")
)

func HttpMethods() []HttpMethod {
	return []HttpMethod{GET, POST, DELETE, HEAD, TRACE, PUT, OPTIONS, CONNECT}
}
