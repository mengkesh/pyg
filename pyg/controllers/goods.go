package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pygGitHub/pyg/models"
	"math"
)

type GoodsController struct {
	beego.Controller
}

//展示首页
func (this *GoodsController) ShowIndex() {
	username := this.GetSession("userName")
	if username != nil {
		this.Data["username"] = username.(string)
	} else {
		this.Data["username"] = ""
	}
	o := orm.NewOrm()
	var oneClass []models.TpshopCategory
	o.QueryTable("TpshopCategory").Filter("Pid", 0).All(&oneClass)
	var types []map[string]interface{}
	//tclass1:=make(map[string]interface{})    不能放在外面，因为map本质是把内存地址存入了切片，如果放在外面，每次的存入的地址都是相同的，那么在最后types里的每个
	//元素都会相同，是tclass1最后一次的值，而如果放在里面，每次tclass1初始化的地址都将不一样，这样types的内容就会不一样。
	for _, v1 := range oneClass {
		tclass1 := make(map[string]interface{})
		tclass1["class11"] = v1
		var twoClass []models.TpshopCategory
		o.QueryTable("TpshopCategory").Filter("Pid", v1.Id).All(&twoClass)
		var tclass2 []map[string]interface{}
		for _, v2 := range twoClass {
			tclass3 := make(map[string]interface{})
			tclass3["class21"] = v2
			var threeClass []models.TpshopCategory
			o.QueryTable("TpshopCategory").Filter("Pid", v2.Id).All(&threeClass)
			tclass3["class22"] = threeClass
			tclass2 = append(tclass2, tclass3)
		}
		tclass1["class12"] = tclass2
		types = append(types, tclass1)
	}
	this.Data["types"] = types
	this.TplName = "index.html"
}

//展示生鲜首页
func (this *GoodsController) ShowIndexSx() {
	o := orm.NewOrm()
	//商品类型
	var goodstypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodstypes)
	//首页轮播
	var indexgoodsbanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexgoodsbanner)
	//促销商品
	var indexpromotionbanner []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&indexpromotionbanner)
	//首页分类商品展示
	var indexshowtypes []map[string]interface{}
	for _, v := range goodstypes {
		indextypes := make(map[string]interface{})
		indextypes["typename"] = v
		var textshow []models.IndexTypeGoodsBanner
		var imgshow []models.IndexTypeGoodsBanner
		qs := o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").Filter("GoodsType", v).OrderBy("Index")
		qs.Filter("DisplayType", 0).All(&textshow)
		qs.Filter("DisplayType", 1).All(&imgshow)
		indextypes["textshow"] = textshow
		indextypes["imgshow"] = imgshow
		indexshowtypes = append(indexshowtypes, indextypes)
	}
	this.Data["indexshowtypes"] = indexshowtypes
	this.Data["indexpromotionbanner"] = indexpromotionbanner
	this.Data["indexgoodsbanner"] = indexgoodsbanner
	this.Data["goodstypes"] = goodstypes
	this.TplName = "index_sx.html"
}

//展示商品详情
func (this *GoodsController) ShowDetail() {
	id, err := this.GetInt("id")
	if err != nil {
		beego.Error(err)
		this.Redirect("/index_sx", 302)
	}
	o := orm.NewOrm()

	//err=o.Read(&goodsSKU)
	//if err!=nil{
	//	beego.Error(err)
	//	this.Redirect("/index_sx",302)
	//}
	//商品类型
	var goodstypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodstypes)
	var goodsSKU models.GoodsSKU
	goodsSKU.Id = id
	err = o.QueryTable("GoodsSKU").RelatedSel("GoodsType", "Goods").Filter("Id", id).One(&goodsSKU)
	if err != nil {
		beego.Error(err)
		this.Redirect("/index_sx", 302)
	}
	var newgoodsSKU []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Name", goodsSKU.GoodsType.Name).OrderBy("-Time").Limit(2, 0).All(&newgoodsSKU)
	this.Data["newgoodsSKU"] = newgoodsSKU
	this.Data["goodstypes"] = goodstypes
	this.Data["goodsSKU"] = goodsSKU
	this.TplName = "detail.html"
}

//页码函数
func Countpages(pagenums, pageindex int) []int {
	var pages []int
	if pagenums <= 5 {
		for i := 1; i <= pagenums; i++ {
			pages = append(pages, i)
		}
	} else if pageindex <= 3 {
		for i := 1; i <= 5; i++ {
			pages = append(pages, i)
		}
	} else if pageindex >= pagenums-2 {
		for i := pagenums - 4; i <= pagenums; i++ {
			pages = append(pages, i)
		}
	} else {
		for i := pageindex - 2; i <= pageindex+2; i++ {
			pages = append(pages, i)
		}
	}
	return pages
}

//列表
func (this *GoodsController) ShowList() {
	id, err := this.GetInt("id")
	if err != nil {
		beego.Error(err)
		this.Redirect("/index_sx", 302)
	}
	o := orm.NewOrm()
	//定每夜显示几项
	singlepagenum := 1
	var goods []models.GoodsSKU
	qs := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id)
	//获取总记录数
	pagecount, _ := qs.Count()
	//获取总页数
	pagenums := int(math.Ceil(float64(pagecount) / float64(singlepagenum)))
	//获取页码
	pageindex, err := this.GetInt("page")
	if err != nil {
		pageindex = 1
	}
	var prepage, nextpage int
	if pageindex <= 1 {
		prepage = 1
	} else {
		prepage = pageindex - 1
	}
	if pageindex >= pagenums {
		nextpage = pagenums
	} else {
		nextpage = pageindex + 1
	}

	//获取页切片
	pages := Countpages(pagenums, pageindex)

	//获取排序依据
	sort := this.GetString("sort")
	if sort == "price" {
		qs = qs.OrderBy("Price")
	} else if sort == "sale" {
		qs = qs.OrderBy("-Sales")
	}
	qs.Limit(singlepagenum, singlepagenum*(pageindex-1)).All(&goods)
	//新品展示
	var newgoodsSKU []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).OrderBy("-Time").Limit(2, 0).All(&newgoodsSKU)
	//查类型
	var goodstypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodstypes)
	this.Data["newgoodsSKU"] = newgoodsSKU
	this.Data["goods"] = goods
	this.Data["goodstypes"] = goodstypes
	this.Data["id"] = id
	this.Data["sort"] = sort
	this.Data["pages"] = pages
	this.Data["prepage"] = prepage
	this.Data["nextpage"] = nextpage
	this.Data["pageindex"] = pageindex
	this.TplName = "list.html"
}

//搜索
func (this *GoodsController) HandleSearch() {
	content := this.GetString("search")
	if content == "" {
		this.Redirect("/index", 302)
		return
	}
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	o.QueryTable("GoodsSKU").Filter("Name__icontains", content).All(&goods)
	this.Data["goods"] = goods
	this.TplName = "search.html"
}
