package controllers

import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.Route("/", Home{})
	wemvc.App.Route("/download", Download{})
}