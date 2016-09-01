package wemvc

import (
	"errors"
	"html/template"
	"bytes"
)

type endRequestError struct {
}

var emptyViewPath = errors.New("the view path canot be empty")

var openDirError = errors.New("dir open err")

func errorViewPath(viewPath string) error {
	return errors.New("cannot find the view path " + viewPath)
}

var errorTpl,_ = template.New("error").Parse(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title>WEMVC {{.Version}} Error Detail - {{.StatusCode}} - {{.Status}}</title>
    <style type="text/css">
        body {margin: 0;font-size: .7em;font-family: Verdana, Arial, Helvetica, sans-serif;}
        code {margin: 0;color: #006600;font-size: 1.1em;font-weight: bold;}
        h1 {font-size: 2.4em;margin: 0;color: #FFF;}
        h2 {font-size: 1.7em;margin: 0;color: #CC0000;}
        h3 {font-size: 1.4em;margin: 10px 0 0 0;color: #CC0000;}
        h4 {font-size: 1.2em;margin: 10px 0 5px 0;}
        pre {margin: 0;font-size: 1.4em;word-wrap: break-word;}
        fieldset{padding:0 15px 10px 15px;word-break:break-all;}
        #content {margin: 0 0 0 2%;position: relative;}
        .summary-container,.content-container{background:#FFF;width:96%;margin-top:8px;padding:10px;position:relative;}
        .content-container p {margin: 0 0 10px 0;}
    </style>
</head>
<body>
    <div id="content">
        <div class="content-container">
            <h3>HTTP Error {{.StatusCode}} - {{.Status}}</h3>
            <hr width="100%" size="1" color="silver">
            <h4>{{.Detail}}</h4>
        </div>
        {{if .DebugStack}}
        <div class="content-container">
            <fieldset>
                <h4>Debug Stack:</h4>
                <code>
                  <pre>{{.DebugStack}}</pre>
                </code>
            </fieldset>
        </div>
        {{end}}
        <div class="content-container">
        	<i>wemvc server version {{.Version}}</i>
        </div>
    </div>
</body>
</html>`)

func renderError(statusCode int, status, errDetail, stack string) []byte {
	var data = make(map[string]interface{})
	data["StatusCode"] = statusCode
	data["Version"] = Version
	data["Status"] = status
	data["Detail"] = errDetail
	if len(stack) > 0{
		data["DebugStack"] = stack
	}
	var buf = &bytes.Buffer{}
	errorTpl.Execute(buf, data)
	return buf.Bytes()
}