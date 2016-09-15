# wemvc
a very simple golang web framework
```
package main

import "github.com/Simbory/wemvc"

type HomeController struct {
	wemvc.Controller
}

func (this HomeController) Index() *wemvc.Result {
	return this.Content("hello world!<br/><a href=\"/about\">About</a>", "text/html")
}

func (this HomeController) GetAbout() *wemvc.Result {
	obj := make(map[string]interface{})
	obj["routeData"] = this.RouteData
	obj["headers"] = this.Request.Header
	return this.JSON(obj)
}

func init() {
	wemvc.Route("/<action=index>", HomeController{})
}

func main() {
	wemvc.Run(8080);
}
```
### another sample
[https://github.com/Simbory/wemvc-sample](https://github.com/Simbory/wemvc-sample)
