package controllers

import (
	"github.com/astaxie/beego"
	"github.com/gomodule/redigo/redis"
	"github.com/astaxie/beego/orm"
	"pygGitHub/pyg/models"
)

type CartController struct {
	beego.Controller
}
//加入购物车
func(this *CartController)HandleAddCart(){
	//获取数据
	id,err:=this.GetInt("goodsId")
	num,err1:=this.GetInt("num")
	resp:=make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	//校验数据
	if err!=nil || err1!=nil{
		resp["errno"] = 1
		resp["errmsg"] = "输入数据不完整"
		return
	}
	name := this.GetSession("userName")
	if name == nil{
		resp["errno"] = 2
		resp["errmsg"] = "当前用户未登录，不能添加购物车"
		return
	}
	//处理数据
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		//this.Redirect("/index_sx",302)
		//return
		resp["errno"] = 3
		resp["errmsg"] = "连接数据库失败，请重试"
		return
	}
	defer conn.Close()
	result,err:=redis.Int(conn.Do("hget","cart_"+name.(string),id))

	_,err=conn.Do("hset","cart_"+name.(string),id,result+num)
	if err!=nil{
		//this.Redirect("/index_sx",302)
		//return
		resp["errno"] = 4
		resp["errmsg"] = "写入数据库失败，请重试"
		return
	}

	//返回数据
        resp["errno"]=5

}
//展示购物车
func(this *CartController)ShowCart(){
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		beego.Error(err)
		this.Redirect("/index_sx",302)
		return
	}
	defer conn.Close()
	name:=this.GetSession("userName")
	result,err:=redis.Ints(conn.Do("hgetall","cart_"+name.(string)))
	o:=orm.NewOrm()
	var goods []map[string]interface{}
	price:=0
	for i:=0;i<len(result);i+=2{
		temp:=make(map[string]interface{})
		var goodsSku models.GoodsSKU
		goodsSku.Id=result[i]
		o.Read(&goodsSku)
		littleprice:=result[i+1]*goodsSku.Price
		temp["count"]=result[i+1]
		temp["goodsSku"]=goodsSku
		temp["littleprice"]=littleprice
		goods=append(goods,temp)
		price+=littleprice
	}
	this.Data["goods"]=goods
	this.Data["price"]=price
	this.Data["num"]=len(goods)
	this.TplName="cart.html"
}
//购物车中添加减少数量
func(this *CartController)HandleChangeCart(){
	count,err1:=this.GetInt("count")
	goodsid,err2:=this.GetInt("goodsid")
	resp:=make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	if err1!=nil || err2!=nil{
		resp["errno"]=1
		resp["errmsg"]="数据传输不完整"
		return
	}
	userName:=this.GetSession("userName")
	if userName==nil{
		resp["errno"]=3
		resp["errmsg"]="用户未登陆"
		return
	}
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errno"]=2
		resp["errmsg"]="数据库连接失败"
		return
	}
	defer conn.Close()
	_,err=conn.Do("hset","cart_"+userName.(string),goodsid,count)
	if err!=nil{
		resp["errno"]=4
		resp["errmsg"]="写入数据库失败"
		return
	}
	resp["errno"]=5
	resp["errmsg"]="OK"
}
func(this *CartController)HandleDeleteCart(){
	goodsid,err:=this.GetInt("goodsid")
	resp:=make(map[string]interface{})
	defer RespFunc(&this.Controller,resp)
	if err!=nil{
		beego.Error(err)
		resp["errno"]=1
		resp["errmsg"]="数据传输不完整"
		return
	}
	userName:=this.GetSession("userName")
	if userName==nil{
		resp["errno"]=2
		resp["errmsg"]="用户未登陆"
		return
	}
	conn,err:=redis.Dial("tcp","127.0.0.1:6379")
	if err!=nil{
		resp["errno"]=3
		resp["errmsg"]="连接数据库失败"
		return
	}
	defer conn.Close()
	_,err=conn.Do("hdel","cart_"+userName.(string),goodsid)
	if err!=nil {
		resp["errno"]=4
		resp["errmsg"]="操作数据库失败"
		return
	}
	resp["errno"]=5
	resp["errmsg"]="OK"
}
