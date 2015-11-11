package controllers
import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.AddRoute("/admin", Index{})
	wemvc.App.AddRoute("/admin/login", Login{})
}