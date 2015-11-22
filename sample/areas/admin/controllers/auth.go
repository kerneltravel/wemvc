package controllers

import "github.com/Simbory/wemvc"

type AuthController struct {
	wemvc.Controller
}

func (this AuthController) OnLoad() {
	loginCookie, err := this.Request.Cookie("ADMIN_AUTH")
	if err != nil || len(loginCookie.Value) < 1 {
		this.Redirect("/admin/account/login?returnUrl=" + this.Request.URL.String())
	}
}
