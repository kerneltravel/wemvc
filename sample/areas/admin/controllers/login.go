package controllers
import "github.com/Simbory/wemvc"

type Login struct {
	wemvc.Controller
}

func (this *Login)Get() wemvc.Response {
	return this.View("admin/login/index")
}

func (this *Login)Post() wemvc.Response {
	this.Request.ParseForm()
	var email = this.Request.Form.Get("email")
	var pwd = this.Request.Form.Get("password")
	if email == "simbory@sina.cn" && pwd == "123456" {
		var returnUrl = this.Request.URL.Query().Get("returnUrl")
		if len(returnUrl) < 1 {
			returnUrl = "/admin"
		}
		return this.Redirect(returnUrl)
	}
	this.ViewData["email"] = email
	this.ViewData["error"] = "invalid email or password"
	return this.View("admin/login/index")
}