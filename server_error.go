package wemvc

import (
	"net/http"
	"runtime/debug"
	"strings"
)

func (app *server) error404(req *http.Request) *ContentResult {
	return renderError(
		404,
		"The resource you are looking for has been removed, had its name changed, or is temporarily unavailable",
		strAdd("Request URL: ", req.URL.String()),
		"",
	)
}

func (app *server) error403(req *http.Request) *ContentResult {
	return renderError(
		403,
		"The server understood the request but refuses to authorize it",
		strAdd("Request URL: ", req.URL.String()),
		"",
	)
}

func (app *server) handleErrorReq(req *http.Request, code int, title ...string) *ContentResult {
	var handler = app.errorHandlers[code]
	if handler != nil {
		return handler(req)
	} else if errTitle, ok := statusCodeMapping[code]; ok {
		t := errTitle
		if len(title) > 0 && len(title[0]) > 0 {
			t = strAdd(t, ":", title[0])
		}
		return renderError(
			code,
			t,
			strAdd("Request URL: ", req.URL.String()),
			"",
		)
	}
	return app.error404(req)
}

func (app *server) panicRecover(res http.ResponseWriter, req *http.Request) {
	rec := recover()
	if rec == nil {
		return
	}
	// detect end request
	_, ok := rec.(*errEndRequest)
	if ok {
		return
	}
	// process 500 error
	res.WriteHeader(500)
	var debugStack string
	if app.config.GetSetting("DebugMode") == "true" {
		debugStack = byte2Str(debug.Stack())
		debugStack = strings.Replace(debugStack, "<", "&lt;", -1)
		debugStack = strings.Replace(debugStack, ">", "&gt;", -1)
	}
	if err, ok := rec.(error); ok {
		res.Write(genError(500, "", err.Error(), debugStack))
	} else {
		res.Write(genError(500, "", "Unkown Internal Server Error", debugStack))
	}
}
