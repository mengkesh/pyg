package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pygGitHub/pyg/models"
)

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
	o:=orm.NewOrm()
	var oneClass []models.TpshopCategory
	o.QueryTable("TpshopCategory").Filter("Pid",0).All(&oneClass)
	var types []map[string]interface{}
	//t:=make(map[string]interface{})    不能放在外面，因为types是个切片，每个值对应一个数组，如果t放在外面，t是不变的，types里的每个元素都会对应到t所在的地址空间的值，那么在最后types里的每个
	//元素都会相同，是t最后一次的值，而如果放在里面，每次t初始化的地址都将不一样，因为types会对应到上次t所在的地址空间的值，这样types的内容就会不一样。
	for _,v:=range oneClass{
		tclass1:=make(map[string]interface{})
		tclass1["class11"]=v
		var twoClass []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid",v.Id).All(&twoClass)
		var tclass2 []map[string]interface{}
		for _,v:=range twoClass{
			tclass3:=make(map[string]interface{})
			tclass3["class21"]=v
			var threeClass []models.TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid",v.Id).All(&threeClass)
			tclass3["class22"]=threeClass
			tclass2=append(tclass2,tclass3)
		}
		tclass1["class12"]=tclass2
		types=append(types,tclass1)
	}
	this.Data["types"]=types
	this.TplName="index.html"
}