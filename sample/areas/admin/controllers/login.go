package controllers
import (
	"github.com/Simbory/wemvc"
	"net/http"
)

type Login struct {
	wemvc.Controller
}

func (this Login) GetLogin() wemvc.Response {
	return this.View("admin/login/index")
}

func (this Login) GetTest() wemvc.Response {
	return this.Content("test test")
}

func (this Login) PostLogin() wemvc.Response {
	var email = this.Request().Form.Get("email")
	var pwd = this.Request().Form.Get("password")
	if email == "simbory@sina.cn" && pwd == "123456" {
		var returnUrl = this.Request().URL.Query().Get("returnUrl")
		if len(returnUrl) < 1 {
			returnUrl = "/admin"
		}
		var cookie = &http.Cookie{
			Name: "ADMIN_AUTH",
			Value: email,
			Path: "/",
			HttpOnly:false,
			Secure: false,
			Domain: this.Request().URL.Host,
		}
		http.SetCookie(this.Response(), cookie)
		return this.Redirect(returnUrl)
	}
	this.ViewData["email"] = email
	this.ViewData["error"] = "invalid email or password"
	return this.View("admin/login/index")
}