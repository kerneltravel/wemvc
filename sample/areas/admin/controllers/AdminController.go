package controllers
import "github.com/Simbory/wemvc"

type AdminController struct {
	wemvc.Controller
}

func (this *AdminController)OnLoad() {
	loginCookie,err := this.Request().Cookie("ADMIN_AUTH")
	if err != nil || len(loginCookie.Value) < 1 {
		this.Redirect("/admin/login?returnUrl=" + this.Request().URL.String())
	}
}