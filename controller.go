package wemvc

import "net/http"

type IController interface {
	Init(*http.Request)
	Get() Response
	Post() Response
	Delete() Response
	Head() Response
	Trace() Response
	Put() Response
	Options() Response
}

type Controller struct {
	Request *http.Request
}

func (this *Controller) Init(req *http.Request) {
	this.Request = req
}

func (this *Controller) Get() Response { return nil }

func (this *Controller) Post() Response { return nil }

func (this *Controller) Delete() Response { return nil }

func (this *Controller) Head() Response { return nil }

func (this *Controller) Trace() Response { return nil }

func (this *Controller) Put() Response { return nil }

func (this *Controller) Options() Response { return nil }

func (this *Controller) Content(str string, ctype ...string) Response {
	var resp = NewResponse()
	resp.Write([]byte(str))
	if len(ctype) > 0 && len(ctype[0]) > 0 {
		resp.SetContentType(ctype[0])
	}
	return resp
}