package controllers

import (
	"github.com/Simbory/wemvc"
	"net/http"
)

type Account struct {
	wemvc.Controller
}

func (this Account) GetLogin() wemvc.ActionResult {
	return this.ViewFile("admin/login/index")
}

func (this Account) PostLogin() wemvc.ActionResult {
	var email = this.Request.Form.Get("email")
	var pwd = this.Request.Form.Get("password")
	if email == "simbory@sina.cn" && pwd == "123456" {
		var returnUrl = this.Request.URL.Query().Get("returnUrl")
		if len(returnUrl) < 1 {
			returnUrl = "/admin"
		}
		var cookie = &http.Cookie{
			Name:     "ADMIN_AUTH",
			Value:    email,
			Path:     "/",
			HttpOnly: false,
			Secure:   false,
			Domain:   this.Request.URL.Host,
		}
		http.SetCookie(this.Response, cookie)
		return this.Redirect(returnUrl)
	}
	this.ViewData["email"] = email
	this.ViewData["error"] = "invalid email or password"
	return this.ViewFile("admin/login/index")
}
