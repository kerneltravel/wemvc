package wemvc

import (
	"bytes"
	"errors"
	"html/template"
	"runtime"
)

type endRequestError struct{}

var emptyViewPathError = errors.New("The view path is empty.")

var openDirError = errors.New("Failed to open the directory.")

var setRootError = errors.New("The web root cannot be change while the application is running.")

var invalidRootError = errors.New("The root directory is invalid.")

var pathPrefixEmptyError = errors.New("The path prefix cannot be empty.")

var invalidPathError = errors.New("The path of the static file cannot be end with '/'")

var invalidNsError = errors.New("The namespace is invalid.")

var resEmptyError = errors.New("The response writer cannot be empty")

var reqEmptyError = errors.New("The http request cannot be empty")

var filterPathError = errors.New("The filter path prefix must starts with '/'")

var notFoundTplError = func(file string) error {
	return errors.New("can't find template file \"" + file + "\"")
}

var sessionProvNilError = errors.New("The session provider is nil")

var sessionRegTwiceError = func(name string) error {
	return errors.New("session: Register called twice for provider " + name)
}

var viewPathNotFoundError = func(viewPath string) error {
	return errors.New("cannot find the view path " + viewPath)
}

var invalidCtrlTypeError = func(typeName string) error {
	return errors.New("Invalid controller type: \"" + typeName + "\"")
}

var ctrlNoActionError = func(typeName string) error {
	return errors.New("Thhe controller \"" + typeName + "\" has no action method")
}

var errorTpl, _ = template.New("error").Parse(`<!DOCTYPE html>
<html>
<head>
    <title>Error {{.StatusCode}} : {{.Status}}</title>
    <meta name="viewport" content="width=device-width" />
    <style>
        body {font-family:"Verdana";font-weight:normal;font-size: .7em;color:black;} 
        p {font-family:"Verdana";font-weight:normal;color:black;margin-top: -5px}
        b {font-family:"Verdana";font-weight:bold;color:black;margin-top: -5px}
        h1 { font-family:"Verdana";font-weight:normal;font-size:18pt;color:red }
        h2 { font-family:"Verdana";font-weight:normal;font-size:14pt;color:maroon }
        pre {font-family:"Consolas","Lucida Console",Monospace;font-size:11pt;margin:0;padding:0.5em;line-height:14pt}
        .marker {font-weight: bold; color: black;text-decoration: none;}
        .version {color: gray;}
        .error {margin-bottom: 10px;}
        .expandable { text-decoration:underline; font-weight:bold; color:navy; cursor:hand; }
        @media screen and (max-width: 639px) {
            pre { width: 440px; overflow: auto; white-space: pre-wrap; word-wrap: break-word; }
        }
        @media screen and (max-width: 479px) {
            pre { width: 280px; }
        }
    </style>
</head>
<body bgcolor="white">
    <span>
        <h1>HTTP Error {{.StatusCode}} : {{.Status}}</h1>
        <hr width=100% size=1 color=silver>
{{if .ErrorTitle}}
		<h2><i>{{.ErrorTitle}}</i></h2>
{{end}}
    </span>
    <font face="Arial, Helvetica, Geneva, SunSans-Regular, sans-serif ">
{{if .ErrorDetail}}
    <b>Error Detail:</b><br><br>
    <table width="100%" bgcolor="#ffffcc">
       <tr>
          <td>
              <code><pre>
{{.ErrorDetail}}</pre></code>
          </td>
       </tr>
    </table>
    <br>
{{end}}
{{if .DebugStack}}
    <b>Debug Stack Trace:</b><br><br>
    <table width="100%" bgcolor="#ffffcc">
       <tr>
          <td>
              <code><pre>
{{.DebugStack}}</pre></code>
          </td>
       </tr>
    </table>
    <br>
{{end}}
    <hr width=100% size=1 color=silver>
    <b>Version:</b>&nbsp;wemvc framework {{.Version}} with {{.GoVersion}}
    </font>
</body>
</html>`)

func renderError(statusCode int, errorTitle, errDetail, stack string) []byte {
	var data = map[string]interface{}{
		"StatusCode":  statusCode,
		"Status":      statusCodeMapping[statusCode],
		"ErrorTitle":  errorTitle,
		"ErrorDetail": errDetail,
    	"Version":     Version,
		"GoVersion":   runtime.Version(),
	}
	if len(stack) > 0 {
		data["DebugStack"] = stack
	}
	var buf = &bytes.Buffer{}
	errorTpl.Execute(buf, data)
	return buf.Bytes()
}
