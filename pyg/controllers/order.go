package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"pygGitHub/pyg/models"
	"strconv"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
	"github.com/smartwalle/alipay"
)

type OrderController struct {
	beego.Controller
}

//展示订单页
func (this *OrderController) ShowOrder() {
	goodids := this.GetStrings("checkGoods")
	if len(goodids) == 0 {
		this.Redirect("/user/cart", 302)
		return
	}
	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var addresses []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName.(string)).All(&addresses)
	for k, v := range addresses {
		qian := v.Phone[:3]
		hou := v.Phone[7:]
		v.Phone = qian + "****" + hou
		addresses[k] = v
	}
	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		beego.Error(err)
		this.Redirect("/user/cart", 302)
		return
	}
	var goods []map[string]interface{}
	order := 0
	totalprice := 0
	for _, v := range goodids {
		good := make(map[string]interface{})
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		o.QueryTable("GoodsSKU").Filter("Id", id).One(&goodsSku)
		good["goodsSku"] = goodsSku
		count, err := redis.Int(conn.Do("hget", "cart_"+userName.(string), id))
		if err != nil {
			beego.Error(err)
			this.Redirect("/user/cart", 302)
			return
		}
		good["count"] = count
		good["littleprice"] = count * goodsSku.Price
		order++
		good["order"] = order
		goods = append(goods, good)
		totalprice += count * goodsSku.Price
	}
	this.Data["goods"] = goods
	this.Data["addresses"] = addresses
	this.Data["goodscount"] = len(goods)
	this.Data["totalprice"] = totalprice
	this.Data["sumtotalprice"] = totalprice + 10
	this.Data["goodids"] = goodids
	this.TplName = "place_order.html"
}

//提交订单
func (this *OrderController) HandlePushOrder() {
	resp := make(map[string]interface{})
	defer RespFunc(&this.Controller, resp)
	userName := this.GetSession("userName")
	if userName == nil {
		beego.Error(1)
		resp["errno"] = 1
		resp["reemsg"] = "请先登陆"
		return
	}
	//获取前端数据
	addrId, err1 := this.GetInt("addrId")
	payId, err2 := this.GetInt("payId")
	goodsIds := this.GetString("goodIds")
	goodsidslice := strings.Split(goodsIds[1:len(goodsIds)-1], " ")
	totalCount, err3 := this.GetInt("totalCount")
	totalPrice, err4 := this.GetFloat("totalPrice")
	if err1 != nil || err2 != nil || len(goodsIds) == 0 || err3 != nil || err4 != nil {
		beego.Error(2)
		resp["errno"] = 2
		resp["reemsg"] = "传输数据不完整"
		return
	}
	var orderinfo models.OrderInfo
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	err := o.Read(&user, "Name")
	if err != nil {
		beego.Error(3)
		resp["errno"] = 3
		resp["reemsg"] = "读取数据库（User）失败"
		return
	}
	var address models.Address
	address.Id = addrId
	err = o.Read(&address)
	if err != nil {
		beego.Error(4)
		resp["errno"] = 4
		resp["reemsg"] = "读取数据库（Address）失败"
		return
	}
	orderinfo.User = &user
	orderinfo.Address = &address
	orderinfo.PayMethod = payId
	orderinfo.TotalCount = totalCount
	orderinfo.TotalPrice = int(totalPrice)
	orderinfo.TransitPrice = 10
	orderinfo.OrderId = time.Now().Format("20060102150405" + strconv.Itoa(user.Id))
	//开启事务
	o.Begin()
	_, err = o.Insert(&orderinfo)
	beego.Info(orderinfo)
	if err != nil {
		beego.Error(9)
		resp["errno"] = 9
		resp["reemsg"] = "插入数据库(OrderInfo)失败"
		o.Rollback()
		return
	}

	conn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		beego.Error(6)
		resp["errno"] = 6
		resp["reemsg"] = "连接redis失败"
		o.Rollback()
		return
	}
	defer conn.Close()
	for _, v := range goodsidslice {
		id, _ := strconv.Atoi(v)
		var goodsSku models.GoodsSKU
		goodsSku.Id = id
		err := o.Read(&goodsSku)
		oldStock := goodsSku.Stock
		//beego.Info("原始库存等于", oldStock)
		if err != nil {
			beego.Error(7)
			resp["errno"] = 7
			resp["reemsg"] = "读取数据库（GoodsSKU）失败"
			o.Rollback()
			return
		}
		count, err := redis.Int(conn.Do("hget", "cart_"+userName.(string), id))
		if err != nil {
			beego.Error(err)
			resp["errno"] = 8
			resp["reemsg"] = "读取redis失败"
			o.Rollback()
			return
		}
		if count > goodsSku.Stock {
			beego.Error(11)
			resp["errno"] = 11
			resp["reemsg"] = "库存不足"
			o.Rollback()
			return
		}
		//goodsSku.Stock-=count
		//goodsSku.Sales+=count

		o.Read(&goodsSku)
		qs := o.QueryTable("GoodsSKU").Filter("Id", id).Filter("Stock", oldStock)
		num, err := qs.Update(orm.Params{"Stock": goodsSku.Stock - count, "Sales": goodsSku.Sales + count})
		if num == 0 {
			resp["errno"] = 12
			resp["errmsg"] = "购买失败，请重新排队！"
			o.Rollback()
			return
		}
		var ordergoods models.OrderGoods
		ordergoods.OrderInfo = &orderinfo
		ordergoods.GoodsSKU = &goodsSku
		ordergoods.Count = count
		ordergoods.Price = goodsSku.Price * count
		_, err = o.Insert(&ordergoods)
		if err != nil {
			beego.Error(10)
			resp["errno"] = 10
			resp["reemsg"] = "插入数据库(OrderGoods)失败"
			o.Rollback()
			return
		}
		_, err = conn.Do("hdel", "cart_"+userName.(string), id)
		if err != nil {
			resp["errno"] = 13
			resp["reemsg"] = "删除购物车失败"
			o.Rollback()
			return
		}
	}
	o.Commit()
	resp["errno"] = 5
	resp["errmsg"] = "OK"
}

/*
addrId=$('input[name=addr]:checked').val()
payId=$('input[name=pay_style]:checked').val()
goodIds=$(this).attr('goodids')
totalCount=$('.common_list_con').find('.total_goods_count').find('em').text()
totalPrice=$('.common_list_con').find('.total_goods_count').find('b').text()
totalCount=parseInt(totalCount)
totalPrice=parseFloat(totalPrice)
params={"addrId":addrId,"payId":payId,"goodIds":goodIds,"totalCount":totalCount,"totalPrice":totalPrice}*/
//付款
func (this *OrderController) HandlePay() {
	orderinfoid, err := this.GetInt("orderinfoid")
	if err != nil {
		this.Redirect("/user/order", 302)
		return
	}
	o := orm.NewOrm()
	var orderinfo models.OrderInfo
	orderinfo.Id = orderinfoid
	err = o.Read(&orderinfo)
	if err != nil {
		this.Redirect("/user/order", 302)
		return
	}
	publiKey := `MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApZcuMlJTIZCsgoXWneUK
XFZw429yzg5+1hguOZSCDrJM+UEQSDh9cSFCfo3bcrnTlVciiDWZ5LTtpzU0Kwz2
FSCf1UZhFQ0pKL1zMwkWrFvj6Q0T5QjJSnEuOmAp6EJLGGEg8hrS/6qnfL7Sg4fz
e87SmC9DaJtNaziA8G7hw4XI0Ri2rKVrOMiuITYcZio9WWtuAgrTGMxhn7dq5vBE
4a6dKMY3EgtWnLjsEBS024D4n6SOyhcfh5k5zyQ7+qrci6Kyaztp1VpbZVSyy0Yj
P/xHm4B29QzvhtlGrwYN7HAXBV8+MC1XPDeKqJ1I5/7VtVcqc9oD+3TX4Hea7bAo
ewIDAQAB`
	privateKey := `MIIEpAIBAAKCAQEApZcuMlJTIZCsgoXWneUKXFZw429yzg5+1hguOZSCDrJM+UEQ
SDh9cSFCfo3bcrnTlVciiDWZ5LTtpzU0Kwz2FSCf1UZhFQ0pKL1zMwkWrFvj6Q0T
5QjJSnEuOmAp6EJLGGEg8hrS/6qnfL7Sg4fze87SmC9DaJtNaziA8G7hw4XI0Ri2
rKVrOMiuITYcZio9WWtuAgrTGMxhn7dq5vBE4a6dKMY3EgtWnLjsEBS024D4n6SO
yhcfh5k5zyQ7+qrci6Kyaztp1VpbZVSyy0YjP/xHm4B29QzvhtlGrwYN7HAXBV8+
MC1XPDeKqJ1I5/7VtVcqc9oD+3TX4Hea7bAoewIDAQABAoIBAQCMTG4YtsW2j6SK
JimzuAayO48END46Ne/jJ7Oql5gmKY2sNiM8fZhTDNIQ9dIz/xw00RHyBDAypdUh
saNUwnORbQXfJvVEZ9Uyrml3mUC7olOU9r6fdHVP/FslnKqFHf4QVuMaHf4eHpIv
0GH+jWaPxKmLxafAbq0GpmYg0GG6TTfthL0TxAd4npX/bDk2kDLx2K0nOrwUZDSM
Kw7RLf/lI1KTrUT+iU4uxvFGN3nPTwecTiR8kxqBpJ7t+mOKSCT4icxBUvvdPhd0
XOH9bk5NPtDquW1UKeJS+qZpZ4xdT4yRuo4iaxwq8i0wZBCrgMmanuRk8ZHghHUQ
oINJT+vJAoGBANDjnalIKGRCToX3NXQ+xro+nlVhYfajOvF8z+ZYBy1GslwHjNAk
OYwOiMnVe9qC7Pm77WxsNYxfYQqtEQiUFfacgV/qA7/gcAA6DBWW1HK26ZBsHjnc
psEa8mT9JnGavEfsaM5YT/HIMUm2+to97yDun5fx0EfKhteGk5QtUimFAoGBAMrv
r5uXquiT1Vxp0XsJqxqt+tAIeyXenTQ338gbuKVmvCvtO9oISUeL+Xa1ZTsA3wVE
yt4aLzMCTszJWy9JAMGxhdozgjCNAKYOzchxrqm4AZGgZDwVOeo4jmqB3aqGO7uB
u5Oxjl8yWkzTqUv94DIgLdQFpoHILjiccFIAsKn/AoGAdG/cK0c4lKJNSOmCl1iC
x8At2+PbinJ0YbWz4W8CGR/GPfxLZp46obJcVz0zu5qtY4t4ja5HrwZffmb4DrMV
BxE4IHG+Q09kvwucPtCDfaotyT4rHw+6t/tAUEC4FC0vdFv4E8UwUtLHfpKLg+lw
CQhaV4UIF2xx+2NdkgQtP00CgYAuQkK6afE4gPJi1XA95q9NLpl8sGI5+KvHCnGF
cOQ/N9LvBG3fPoJNv9eGusSvlXxA/DRuOnPF4eHKhp+1gKOeg3PqkFE99fZO5BL+
fQN+hoY9Bt2yYHhKLsgv+RhpVZ3qGSGEAjZc9uJknt75ho6DfphTu1IARXxbxTVJ
TAT5SwKBgQDCds2anasxr4gnalzGMfcDezXdladglTXGwUqAiimrxQNFHXepGg6Q
EaDCMxwuM/Y7gz17K7Yshj7I7MsnSV5RY4OQ25wQCsG3xiI3ZMKheQSP1UsVKlAu
NPpHzsLFKCFMzWduC3X60Sc6Fp0dHS8e4baH3H4/dC9SnQ6N5eDSSA==`
	client := alipay.New("2016093000634178", publiKey, privateKey, false)
	var p = alipay.TradePagePay{}
	p.NotifyURL = "http://192.168.168.183:8080/payOK"
	p.ReturnURL = "http://192.168.168.183:8080/payOK"
	p.Subject = "品优购"
	p.OutTradeNo = orderinfo.OrderId
	p.TotalAmount = strconv.Itoa(orderinfo.TotalPrice)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, err := client.TradePagePay(p)
	if err != nil {
		beego.Error("支付失败")
	}
	payUrl := url.String()
	this.Redirect(payUrl, 302)
}
func (this OrderController) HandlePayOk() {
	orderid := this.GetString("out_trade_no")
	tradeno:=this.GetString("trade_no")
	if orderid == "" || tradeno==""{
		beego.Error("出错")
	this.Redirect("/user/order",302)
	return
	}
	o:=orm.NewOrm()
	var orderinfo models.OrderInfo
	orderinfo.OrderId=orderid
	_,err:=o.QueryTable("OrderInfo").Filter("OrderId",orderid).Update(orm.Params{"Orderstatus": 1, "TradeNo": tradeno})
	if err!=nil{
		beego.Error(err)
		this.Redirect("/user/order",302)
	}
	this.Redirect("/user/order",302)
}
