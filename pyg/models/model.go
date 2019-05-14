package models

import (
	"github.com/astaxie/beego/orm"
	_"github.com/go-sql-driver/mysql")

type User struct {
	Id int
	Name string `orm:"size(40);unique"`
	Pwd string `orm:"size(40)"`
	Phone string `orm:"size(11)"`
	Email string `orm:"null"`
	Active bool `orm:"default(false)"`
	Address []*Address `orm:"reverse(many)"`
}
type Address struct {
	Id int
	Receiver string `orm:"size(40)"`
	Addr string `orm:"size(100)"`
	IsDefault bool `orm:"default(false)"`
	PostCode string
	Phone string `orm:"size(11)"`
	User *User `orm:"rel(fk)"`
}
type TpshopCategory struct {
	Id int
	CateName string `orm:"default('')"`
	Pid int `orm:"default(0)"`
	IsShow int `orm:"default(1)"`
	CreateTime int `orm:"null"`
	UpdateTime int `orm:"null"`
	DeleteTime int `orm:"null"`
}
func init(){
	orm.RegisterDataBase("default","mysql","root:123456@tcp(127.0.0.1:3306)/pyg?charset=utf8&loc=Asia%2FShanghai")

	orm.RegisterModel(new(User),new(Address),new(TpshopCategory))

	orm.RunSyncdb("default",false,true)

}
