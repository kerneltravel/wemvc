package main

import (
	_"github.com/Simbory/wemvc"
	_"github.com/Simbory/wemvc/sample/controllers"
	_ "github.com/Simbory/wemvc/sample/areas/admin/controllers"
	"github.com/Simbory/wemvc"
)

func main() {
	println("************************************************************")
	println("*   The web application is started...")
	println("************************************************************")
	wemvc.App.SetStaticPath("/css/")
	wemvc.App.SetStaticPath("/js/")
	wemvc.App.SetStaticPath("/favicon.ico")
	wemvc.App.Run()
	println("************************************************************")
}