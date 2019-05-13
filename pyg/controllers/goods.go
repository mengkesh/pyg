package controllers

import "github.com/astaxie/beego"

type GoodsController struct {
	beego.Controller
}
//展示首页
func(this *GoodsController)ShowIndex(){
	username:=this.GetSession("userName")
	if username!=nil {
		this.Data["username"]=username.(string)
	}else{
		this.Data["username"]=""
	}
	this.TplName="index.html"
}