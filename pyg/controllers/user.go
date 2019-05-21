package controllers

import (
	"github.com/astaxie/beego"
	"regexp"
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"math/rand"
	"time"

	"fmt"
	"github.com/astaxie/beego/orm"
	"pygGitHub/pyg/models"
	"strings"
	"github.com/astaxie/beego/utils"
	"encoding/base64"
	"math"
)

type Message struct {
	Message   string
	RequestId string
	BizId     string
	Code      string
}
type UserController struct {
	beego.Controller
}

func (this *UserController) ShowRegister() {
	this.TplName = "register.html"

}

//封装一个返回json的函数
func RespFunc(this *beego.Controller, resp map[string]interface{}) {
	//3.把容器传给前端
	this.Data["json"] = resp
	//4.指定传递方式
	this.ServeJSON()
}

//发送短信
func (this *UserController) HandleSendMsg() {
	phone := this.GetString("phone")
	//1.定义一个传递给ajax json数据的容器
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)
	//检查手机号是否空
	if phone == "" {
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 1
		resp["errmsg"] = "获取电话号码错误"
		return
	}
	reg, _ := regexp.Compile(`^1[3-9][0-9]{9}$`)
	result := reg.FindString(phone)
	//检查手机号格式是否正确
	if result == "" {
		beego.Error("获取电话号码失败")
		//2.给容器赋值
		resp["errno"] = 2
		resp["errmsg"] = "获取电话号码格式错误"
		return
	}
	rand.Seed(time.Now().UnixNano())

	vcode := fmt.Sprintf("%d", (rand.Intn(9)+1)*100000+(rand.Intn(9)+1)*10000+(rand.Intn(9)+1)*1000+(rand.Intn(9)+1)*100+
		(rand.Intn(9)+1)*10+(rand.Intn(9)+1))

	//发送短信   SDK调用
	client, err := sdk.NewClientWithAccessKey("default", "LTAI3yPOlEWwd1FS", "iH61UaC4fwVCex0tqYLq17hB3S7GbF")
	if err != nil {
		beego.Error(err)
		//2.给容器赋值
		resp["errno"] = 3
		resp["errmsg"] = "初始化短信错误"
		return
	}

	request := requests.NewCommonRequest()
	request.Method = "POST"
	request.Scheme = "https" // https | http
	request.Domain = "dysmsapi.aliyuncs.com"
	request.Version = "2017-05-25"
	request.ApiName = "SendSms"
	request.QueryParams["PhoneNumbers"] = phone
	request.QueryParams["SignName"] = "品优购"
	request.QueryParams["TemplateCode"] = "SMS_165117312"
	request.QueryParams["TemplateParam"] = "{\"code\":" + vcode + "}"

	response, err := client.ProcessCommonRequest(request)
	if err != nil {
		beego.Error(err)
		//2.给容器赋值
		resp["errno"] = 4
		resp["errmsg"] = err
		return
	}

	//json数据解析
	var message Message
	json.Unmarshal(response.GetHttpContentBytes(), &message)
	if message.Message != "OK" {
		beego.Error("电话号码格式错误")
		//2.给容器赋值
		resp["errno"] = 6
		resp["errmsg"] = message.Message
		return
	}

	resp["errno"] = 5
	resp["errmsg"] = "发送成功"
	resp["code"] = vcode
}

//处理注册
func (this *UserController) HandleRegister() {
	phone := this.GetString("phone")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	if phone == "" || pwd == "" || rpwd == "" {
		beego.Error("获取数据错误")
		this.Data["errmsg"] = "获取数据错误"
		this.TplName = "register.html"
		return
	}
	if pwd != rpwd {
		beego.Error("两次密码输入不一致")
		this.Data["errmsg"] = "两次密码输入不一致"
		this.TplName = "register.html"
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = phone
	user.Pwd = pwd
	user.Phone = phone
	o.Insert(&user)
	this.Ctx.SetCookie("userName", user.Name, 60*10)
	this.Redirect("/register-email", 302)
}

//展示邮箱激活
func (this *UserController) ShowEmail() {
	this.TplName = "register-email.html"
}

//处理邮箱激活
func (this *UserController) HandleEmail() {
	//获取数据
	email := this.GetString("email")
	pwd := this.GetString("password")
	rpwd := this.GetString("repassword")
	//校验数据
	if email == "" || pwd == "" || rpwd == "" {
		beego.Error("输入数据不完整")
		this.Data["errmsg"] = "输入数据不完整"
		this.TplName = "register-email.html"
		return
	}
	//两次密码是否一直
	if pwd != rpwd {
		beego.Error("两次密码输入不一致")
		this.Data["errmsg"] = "两次密码输入不一致"
		this.TplName = "register-email.html"
		return
	}
	emaillow := strings.ToLower(email)
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(emaillow)
	if result == "" {
		beego.Error("邮箱格式错误")
		this.Data["errmsg"] = "邮箱格式错误"
		this.TplName = "register-email.html"
		return
	}
	beego.Info(result)
	//处理数据
	//发送邮件
	config := `{"username":"houbo940827@163.com","password":"heimago3","host":"smtp.163.com","port":25}`
	emailreg := utils.NewEMail(config)
	emailreg.Subject = "品优购用户激活"
	emailreg.From = "houbo940827@163.com"
	emailreg.To = []string{email}
	userName := this.Ctx.GetCookie("userName")
	emailreg.HTML = `<a href="http://192.168.168.183:8080/active?username=` + userName + `">点击激活</a>`
	emailreg.Send()
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error("用户名错误")
		this.Redirect("/register-email", 302)
		return
	}
	user.Email = email
	o.Update(&user, "Email")

	this.Redirect("/login",302)
}

//激活
func (this *UserController) Active() {
	username := this.GetString("username")
	o := orm.NewOrm()
	var user models.User
	user.Name = username
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error(err)
		this.Redirect("/register-email", 302)
		return
	}
	user.Active = true
	o.Update(&user, "Active")
	this.Redirect("/login", 302)
}

//展示登陆页面
func (this *UserController) ShowLogin() {
	username := this.Ctx.GetCookie("usernameremember")
	if username != "" {
		dec, _ := base64.StdEncoding.DecodeString(username)
		this.Data["nsername"] = string(dec)
		this.Data["check"] = "checked"
	} else {
		this.Data["nsername"] = ""
		this.Data["check"] = ""
	}
	this.TplName = "login.html"
}

//登陆
func (this *UserController) Login() {
	username := this.GetString("nsername")
	pwd := this.GetString("pwd")
	remember := this.GetString("m1")
	beego.Info(remember)
	if username == "" || pwd == "" {
		beego.Error("输入信息不完整")
		this.Redirect("/login", 302)
		return
	}
	o := orm.NewOrm()
	var user models.User
	//user.Name=username
	//err:=o.Read(&user,"Name")
	//if err!=nil{
	//	err:=o.Read(&user,"Phone")
	//	if err!=nil{
	//		err:=o.Read(&user,"Email")
	//		if err!=nil{
	//			beego.Error("用户名不存在")
	//			this.Redirect("/login",302)
	//			return
	//		}
	//	}
	//}
	//if user.Pwd!=pwd{
	//	beego.Error("密码输入错误")
	//	this.Redirect("/login",302)
	//	return
	//}
	reg, _ := regexp.Compile(`^\w[\w\.-]*@[0-9a-z][0-9a-z-]*(\.[a-z]+)*\.[a-z]{2,6}$`)
	result := reg.FindString(username)
	if result != "" {
		user.Email = username
		err := o.Read(&user, "Email")
		if err != nil {
			beego.Error(err)
			this.Data["errmsg"] = "邮箱未注册"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			beego.Error("密码错误")
			this.Data["errmsg"] = "密码错误"
			this.TplName = "login.html"
			return
		}

	} else {
		user.Name = username
		err := o.Read(&user, "Name")
		if err != nil {
			beego.Error(err)
			this.Data["errmsg"] = "用户名不存在"
			this.TplName = "login.html"
			return
		}
		if user.Pwd != pwd {
			beego.Error("密码错误")
			this.Data["errmsg"] = "密码错误"
			this.TplName = "login.html"
			return
		}

	}

	//校验用户是否激活
	if user.Active == false {
		beego.Error("未激活")
		this.Data["errmsg"] = "当前用户未激活，请去目标邮箱激活！"
		this.TplName = "login.html"
		return
	}

	if remember == "on" {
		enc := base64.StdEncoding.EncodeToString([]byte(user.Name))
		this.Ctx.SetCookie("usernameremember", enc, 600)
	} else {
		this.Ctx.SetCookie("usernameremember", user.Name, -1)
	}
	this.SetSession("userName", user.Name)
	this.Redirect("/index", 302)
}

//退出
func (this *UserController) HandleLogout() {
	this.DelSession("userName")
	this.Redirect("/login", 302)
}

//展示用户中心页
func (this *UserController) ShowUserCenterInfo() {
	this.Data["yellowpoint"]=1
	username:=this.GetSession("userName").(string)
	o:=orm.NewOrm()
	var user models.User
	user.Name=username
	err:=o.Read(&user,"Name")
	if err != nil {
		beego.Error(err)
		this.Redirect("/user/usercenterinfo",302)
		return
	}
	var address models.Address
	qs:=o.QueryTable("Address").RelatedSel("User").Filter("User__Name",username)
	err=qs.Filter("Phone",user.Phone).Filter("IsDefault",true).One(&address)
	if err != nil {
		err:=qs.Filter("IsDefault",true).One(&address)
		if err != nil {
			this.Data["address"]=""
		}
	}
	this.Data["address"]=address.Addr
	this.Data["user"]=user
	this.Data["username"]=username
	this.Layout = "layout.html"
	this.TplName = "user_center_info.html"
}

//展示用户中心订单页
func (this *UserController) ShowOrder() {
	username:=this.GetSession("userName").(string)
	singlepagenum:=2
	o:=orm.NewOrm()
	count,_:=o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name",username).Count()
	pagenums:=int(math.Ceil(float64(count)/float64(singlepagenum)))
	pageindex,err:=this.GetInt("pageindex")
	if err!=nil{
		pageindex=1
	}
	prepage:=pageindex-1
	if prepage<=1 {
		prepage=1
	}
	nextpage:=pageindex+1
	if nextpage>= pagenums{
		nextpage=pagenums
	}

	pages := Countpages(pagenums, pageindex)
	var orderinfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Name",username).OrderBy("-Time").Limit(singlepagenum,singlepagenum*(pageindex-1)).All(&orderinfos)
	var orderinfolist []map[string]interface{}
	for _,v:=range orderinfos{
		oderinfocontent:=make(map[string]interface{})
		oderinfocontent["orderinfo"]=v
		var ordergoods []models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("GoodsSku","OrderInfo").Filter("OrderInfo",v).All(&ordergoods)
		oderinfocontent["ordergoods"]=ordergoods
		orderinfolist=append(orderinfolist,oderinfocontent)
	}
	this.Data["pageindex"]=pageindex
	this.Data["prepage"]=prepage
	this.Data["nextpage"]=nextpage
	this.Data["pages"]=pages
	this.Data["orderinfolist"]=orderinfolist
	//this.Data["yellowpoint"]=2
	this.Data["username"]=username
	this.Layout = "layout.html"
	this.TplName = "user_center_order.html"
}

//展示用户中心地址页
func (this *UserController) ShowSite() {
	this.Data["yellowpoint"]=3
	username := this.GetSession("userName")
	name := username.(string)
	o := orm.NewOrm()
	var address models.Address
	//beego.Info(address)
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", name)
	err := qs.Filter("IsDefault", true).One(&address)
	if err != nil {
		//beego.Info(address)
		this.Data["address"] = nil
	} else {
		this.Data["address"] = address
		qian :=address.Phone[:3]
		hou := address.Phone[7:]
		phonehandle:= qian + "****" + hou
		this.Data["phone"] = phonehandle
	}
	this.Data["username"]=name
	this.Layout = "layout.html"
	this.TplName = "user_center_site.html"
}

//处理添加地址
func (this *UserController) AddSite() {
	username := this.GetSession("userName")
	postname := this.GetString("postname")
	postsite := this.GetString("postsite")
	postcode := this.GetString("postcode")
	postphone := this.GetString("postphone")
	if postname == "" || postsite == "" || postcode == "" || postphone == "" {
		beego.Error("输入的信息不完整")
		this.Redirect("/user/site", 302)
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = username.(string)
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error(err)
		this.Redirect("/user/site", 302)
		return
	}
	var address models.Address
	address.Receiver = postname
	address.Addr = postsite
	address.PostCode = postcode
	address.Phone = postphone
	address.User = &user
	//判断是否有默认地址
	var oldaddress models.Address
	qs := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", user.Name)
	err = qs.Filter("IsDefault", true).One(&oldaddress)
	if err == nil {
		oldaddress.IsDefault = false
		o.Update(&oldaddress, "IsDefault")
	}
	address.IsDefault = true
	//插入address表
	_, err = o.Insert(&address)
	if err != nil {
		beego.Error(err)
		this.Redirect("/user/site", 302)
		return
	}
	this.Redirect("/user/site", 302)
}
