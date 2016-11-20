package wemvc

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// CtxFilter request filter func
type CtxFilter func(ctx *Context)

var (
	dangerChars = []string{"<", ">", "&", "%", "*"}
)

func dangerCheck(ctx *Context) {
	for _, c := range dangerChars {
		var ltIndex = strings.Index(ctx.Request().URL.Path, c)
		if ltIndex >= 0 {
			panic(fmt.Errorf("The dangerous character '%s' found in the request path: %d", c, ltIndex))
		}
	}
}

// serveStatic serve the current request as static request
func serveStatic(ctx *Context) {
	physicalFile := ""
	var f = ctx.app.mapPath(ctx.req.URL.Path)
	stat, err := os.Stat(f)
	if err == nil {
		if stat.IsDir() {
			absolutePath := ctx.req.URL.Path
			if !strings.HasSuffix(absolutePath, "/") {
				absolutePath = strAdd(absolutePath, "/")
			}
			physicalPath := ctx.app.mapPath(absolutePath)
			if IsDir(physicalPath) {
				var defaultUrls = ctx.app.config.GetDefaultUrls()
				if len(defaultUrls) > 0 {
					for _, f := range defaultUrls {
						var file = ctx.app.mapPath(strAdd(absolutePath, f))
						if IsFile(file) {
							physicalFile = file
							break
						}
					}
				}
			}
		} else {
			physicalFile = f
		}
		if len(physicalFile) > 0 {
			ctx.Result = &FileResult{FilePath: physicalFile}
		}
	}
	ctx.EndContext()
}

// HandleRouteTree handle request route tree
func handleRoute(ctx *Context) {
	if ctx.Route == nil {
		ctx.Route = &CtxRoute{}
	}
	if len(ctx.Route.RouteUrl) == 0 {
		ctx.Route.RouteUrl = ctx.Request().URL.Path
	}

	var urlPath = ctx.Route.RouteUrl
	if len(urlPath) > 1 && strings.HasSuffix(urlPath, "/") {
		urlPath = strings.TrimRight(urlPath, "/")
	}
	//var resp ActionResult
	cInfo, routeData, err := ctx.app.routing.lookup(ctx.Route.RouteUrl, strings.ToLower(ctx.req.Method))
	if err == nil && cInfo != nil {
		if routeData == nil {
			routeData = make(map[string]string)
		}
		var action = routeData["action"]
		var ns = cInfo.NsName
		if len(action) == 0 {
			action = cInfo.DefaultAction
		} else {
			action = strings.Replace(action, "-", "_", -1)
		}
		var method = strings.ToLower(ctx.req.Method)
		// find the action method in controller
		if ok, actionMethod := cInfo.containsAction(action, method); ok {
			ctx.Route.NsName = ns
			ctx.Ctrl = &CtxController{
				ControllerName: cInfo.CtrlName,
				ControllerType: cInfo.CtrlType,
				ActionName:       action,
				ActionMethodName: actionMethod,
			}
			routeData["controller"] = ctx.Ctrl.ControllerName
			ctx.Route.RouteData = routeData
			return
		}
	}
	if err != nil {
		ctx.Result = ctx.app.handleErrorReq(ctx.Request(), 500, err.Error())
		return
	}
	ctx.EndContext()
}

func execFilters(ctx *Context) {
	if ctx == nil || ctx.Route == nil || ctx.Ctrl == nil {
		return
	}
	urlPath := ctx.Route.RouteUrl
	if !ctx.app.routing.MatchCase {
		urlPath = strings.ToLower(urlPath)
	}
	if len(ctx.Route.NsName) < 1 {
		ctx.app.execFilters(urlPath, ctx)
	} else {
		ns, ok := ctx.app.namespaces[ctx.Route.NsName]
		if ok && ns != nil {
			ns.execFilters(urlPath, ctx)
		}
	}
}

func execAction(ctx *Context) {
	if ctx == nil || ctx.Route == nil || ctx.Ctrl == nil {
		return
	}
	var ctrl = reflect.New(ctx.Ctrl.ControllerType)
	// validate action method
	ctx.Ctrl.ActionMethod = ctrl.MethodByName(ctx.Ctrl.ActionMethodName)
	if !ctx.Ctrl.ActionMethod.IsValid() {
		return
	}
	// call OnInit method
	onInitMethod := ctrl.MethodByName("OnInit")
	if onInitMethod.IsValid() {
		onInitMethod.Call([]reflect.Value{
			reflect.ValueOf(ctx),
		})
	}
	//parse form
	if ctx.req.Method == "POST" || ctx.req.Method == "PUT" || ctx.req.Method == "PATCH" {
		if ctx.req.MultipartForm != nil {
			var size int64
			var maxSize = ctx.app.config.GetSetting("MaxFormSize")
			if len(maxSize) < 1 {
				size = 10485760
			} else {
				size, _ = strconv.ParseInt(maxSize, 10, 64)
			}
			ctx.req.ParseMultipartForm(size)
		} else {
			ctx.req.ParseForm()
		}
	}
	// call action method
	values := ctx.Ctrl.ActionMethod.Call(nil)
	if len(values) == 1 {
		ctx.Result = values[0].Interface()
	}
}
