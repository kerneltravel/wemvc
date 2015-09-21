package controllers
import "github.com/Simbory/wemvc"

func init() {
	wemvc.App.AddRoute("/", &Home{})
}