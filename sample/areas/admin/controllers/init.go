package controllers
import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.Route("/admin", Index{})
	wemvc.App.Route("/admin/login", Login{})
}