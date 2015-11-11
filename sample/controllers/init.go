package controllers
import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.AddRoute("/", Home{})
	wemvc.App.AddRoute("/download", Download{})
	wemvc.App.AddRoute("/download/{action}", Download{})
}