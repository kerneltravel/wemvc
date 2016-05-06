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
	obj["viewData"] = this.RouteData
	obj["headers"] = this.Request.Header
	return this.JSON(obj)
}

func init() {
	wemvc.App.Route("/", HomeController{})
	wemvc.App.Route("/:action", HomeController{})
}

func main() {
	wemvc.App.Port(8080);
	wemvc.App.Run();
}
```
