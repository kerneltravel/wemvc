package main

import "github.com/Simbory/wemvc"
import _ "github.com/Simbory/wemvc/sample/controllers"
import _ "github.com/Simbory/wemvc/sample/areas/admin/controllers"

func main() {
	println("************************************************************")
	println("*   The web application is started...")
	println("************************************************************")
	wemvc.App.Run()
	println("************************************************************")
}
