package controllers

import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.Route("/", Home{})
	wemvc.App.Route("/download", Download{})
	wemvc.App.Route("/test", TestController{})

	wemvc.App.SetFilter("/", func(ctx wemvc.Context) {
		ctx.SetItem("name", "Simbory")
	})
}