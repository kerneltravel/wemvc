# wemvc
a very simple golang web framework
```
package main

import "github.com/Simbory/wemvc"

type HomeController struct {
	wemvc.Controller
}

func (this HomeController) Index() wemvc.ActionResult {
	return this.Content("hello world!<br/><a href=\"/about\">About</a>", "text/html")
}

func (this HomeController) GetAbout() wemvc.ActionResult {
	obj := make(map[string]interface{})
	obj["routeData"] = this.RouteData
	obj["headers"] = this.Request.Header
	return this.Json(obj)
}

func init() {
	wemvc.Route("/", HomeController{})
	wemvc.Route("/:action", HomeController{})
}

func main() {
	wemvc.Run(8080);
}
```
### another sample
[https://github.com/Simbory/wemvc-sample](https://github.com/Simbory/wemvc-sample)