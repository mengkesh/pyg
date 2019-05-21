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
    //展示生鲜首页
    beego.Router("/index_sx",&controllers.GoodsController{},"get:ShowIndexSx")
    //展示商品细节
    beego.Router("/goodsdetail",&controllers.GoodsController{},"get:ShowDetail")
    //商品列表
    beego.Router("/goodsType",&controllers.GoodsController{},"get:ShowList")
    //搜索
    beego.Router("/sreach",&controllers.GoodsController{},"post:HandleSearch")
    //添加购物车
    beego.Router("/addcart",&controllers.CartController{},"post:HandleAddCart")
    //展示购物车
    beego.Router("/user/cart",&controllers.CartController{},"get:ShowCart")
    //购物车中添加减少
    beego.Router("/changeCart",&controllers.CartController{},"post:HandleChangeCart")
    //购物车中删除
    beego.Router("/deleteCart",&controllers.CartController{},"post:HandleDeleteCart")
    //订单页
    beego.Router("/user/addOrder",&controllers.OrderController{},"post:ShowOrder")
    //提交订单
    beego.Router("/pushOrder",&controllers.OrderController{},"post:HandlePushOrder")
    //支付
    beego.Router("/pay",&controllers.OrderController{},"get:HandlePay")
    //支付成功
    beego.Router("/payOK",&controllers.OrderController{},"get:HandlePayOk")
}
func Filterlogin(ctx *context.Context){
	username:=ctx.Input.Session("userName")
	if username==nil{
		ctx.Redirect(302,"/login")
		return
	}
}
