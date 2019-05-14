package routers

import (
	"pygGitHub/pyg/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*",beego.BeforeExec,Filterlogin)
    beego.Router("/", &controllers.MainController{})
    beego.Router("/register",&controllers.UserController{},"get:ShowRegister;post:HandleRegister")
    beego.Router("/sendMsg",&controllers.UserController{},"post:HandleSendMsg")
    beego.Router("/register-email",&controllers.UserController{},"get:ShowEmail;post:HandleEmail")
    beego.Router("/active",&controllers.UserController{},"get:Active")
    beego.Router("/login",&controllers.UserController{},"get:ShowLogin;post:Login")
    beego.Router("/user/logout",&controllers.UserController{},"get:HandleLogout")
    beego.Router("/index",&controllers.GoodsController{},"get:ShowIndex")
    //用户中心
    beego.Router("/user/usercenterinfo",&controllers.UserController{},"get:ShowUserCenterInfo")
    //用户中心地址页
    beego.Router("/user/site",&controllers.UserController{},"get:ShowSite;post:AddSite")
    //展示订单页
    beego.Router("/user/order",&controllers.UserController{},"get:ShowOrder")
}
func Filterlogin(ctx *context.Context){
	username:=ctx.Input.Session("userName")
	if username==nil{
		ctx.Redirect(302,"/login")
		return
	}
}
