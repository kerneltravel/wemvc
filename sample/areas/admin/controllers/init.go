package controllers

import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.Route("/admin", Admin{})
	wemvc.App.Route("/admin/account/:action", Account{})
}
