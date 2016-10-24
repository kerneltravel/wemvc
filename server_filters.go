package wemvc

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

// CtxFilter request filter func
type CtxFilter func(ctx *Context)

// ServeStatic serve static request function
func ServeStatic(ctx *Context) {
	if !ctx.app.isStaticRequest(ctx.req) {
		return
	}
	physicalFile := ""
	var f = ctx.app.MapPath(ctx.req.URL.Path)
	stat, err := os.Stat(f)
	if err == nil {
		if stat.IsDir() {
			absolutePath := ctx.req.URL.Path
			if !strings.HasSuffix(absolutePath, "/") {
				absolutePath = absolutePath + "/"
			}
			physicalPath := ctx.app.MapPath(absolutePath)
			if IsDir(physicalPath) {
				var defaultUrls = ctx.app.config.getDefaultUrls()
				if len(defaultUrls) > 0 {
					for _, f := range defaultUrls {
						var file = ctx.app.MapPath(absolutePath + f)
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

// InitRoute init route request data function
func InitRoute(ctx *Context) {
	if ctx == nil {
		return
	}
	ctx.Route = &CtxRoute{
		RouteUrl: ctx.Request().URL.Path,
	}
}

// HandleRouteTree handle request route tree
func HandleRoute(ctx *Context) {
	if ctx == nil || ctx.Route == nil {
		return
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
		if len(action) < 1 {
			action = cInfo.DefaultAction
		} else {
			action = strings.Replace(action, "-", "_", -1)
		}
		var method = strings.ToLower(ctx.req.Method)
		// find the action method in controller
		if ok, actionMethod := cInfo.containsAction(action, method); ok {
			ctx.Route.NsName = ns
			ctx.Ctrl = &CtxController{
				ControllerName: getContrllerName(cInfo.CtrlType),
				ControllerType: cInfo.CtrlType,

				ActionName:       action,
				ActionMethodName: actionMethod,
			}
			routeData["controller"] = ctx.Ctrl.ControllerName
			ctx.Route.RouteData = routeData
			return
		} else {
			cInfo = nil
		}
	}
	if err != nil {
		ctx.Result = ctx.app.handleError(ctx.Request(), 500, err.Error())
		return
	}
	ctx.EndContext()
}

func ExecutePathFilters(ctx *Context) {
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

func ExecuteAction(ctx *Context) {
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