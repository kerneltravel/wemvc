package wemvc

const (
	Version = "0.1"
	GET     = "GET"
	POST    = "POST"
	DELETE  = "DELETE"
	HEAD    = "HEAD"
	TRACE   = "TRACE"
	PUT     = "PUT"
	OPTIONS = "OPTIONS"
	CONNECT = "CONNECT"
)

func HttpMethods() []string {
	return []string{GET, POST, DELETE, HEAD, TRACE, PUT, OPTIONS, CONNECT}
}
