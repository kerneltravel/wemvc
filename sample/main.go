package main

import "github.com/Simbory/wemvc"
import _ "github.com/Simbory/wemvc/sample/controllers"

func main() {
	println("************************************************************")
	println("*   The web application is started...")
	println("************************************************************")
	wemvc.App.Run()
	println("************************************************************")
}
