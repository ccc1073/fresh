package controllers

import (
	"fresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"strconv"
)

type CartController struct {
	beego.Controller
}

func GetCartCount(this *beego.Controller) {

}

// 添加购物车
func (this *CartController) HandleAddcart() {
	skuid, err := this.GetInt("skuid")
	count, err2 := this.GetInt("count")

	resp := make(map[string]interface{})

	if err != nil || err2 != nil {
		resp["code"] = 400
		resp["errmsg"] = "传递数据不正确"
		this.Data["json"] = resp
		return
	}
	defer this.ServeJSON()

	userName := this.GetSession("userName")

	if userName == nil {
		resp["code"] = 400
		resp["errmsg"] = "用户未登录"
		this.Data["json"] = resp
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "6379")

	if err != nil {
		resp["code"] = 500
		resp["errmsg"] = "redis连接错误"
		this.Data["json"] = resp
		return
	}
	conn.Do("hset", "cart_"+strconv.Itoa(user.Id), skuid, count)

	resp["code"] = 200
	resp["msg"] = "ok"
	this.Data["json"] = resp

}

//展示购物车
func (this *CartController) ShowCart() {

	userName := GetUser(&this.Controller)

	// 从redis中获取数据
	conn, err := redis.Dial("tcp", "6379")
	if err != nil {
		return
	}
	defer conn.Close()

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	resp, err := conn.Do("hgetall", "cart_"+strconv.Itoa(user.Id)) // 返回一个[]map[string]interface

	goodsMap, _ := redis.IntMap(resp, err)

	goods := make([]map[string]interface{}, len(goodsMap))
	i := 0
	totalPrice := 0
	totalCount := 0
	for index, value := range goodsMap {

		skuid, _ := strconv.Atoi(index)
		var goodsSku models.GoodsSKU
		goodsSku.Id = skuid
		o.Read(&goodsSku)

		temp := make(map[string]interface{})

		temp["goodssku"] = goodsSku
		temp["count"] = value

		totalCount += value
		totalPrice += goodsSku.Price * value

		temp["addPrice"] = goodsSku.Price * value

		goods[i] = temp

		i += 1

	}

	this.Data["totalPrice"] = totalPrice
	this.Data["totalCount"] = totalCount
	this.Data["goods"] = goods

	this.TplName = "cart.html"

}

// 更新购物车数据
func (this *CartController) HandleUpdateCart() {

	skuid, err := this.GetInt("skuid")

	count, err2 := this.GetInt("count")

	resp := make(map[string]interface{})

	defer this.ServeJSON()

	if err != nil || err2 != nil {
		resp["code"] = 400
		resp["errmsg"] = "请求数据不正确"
		this.Data["json"] = resp
		return
	}
	userName := this.GetSession("userName")
	if userName == nil {
		resp["code"] = 400
		resp["errmsg"] = "用户未登录"
		this.Data["json"] = resp
		return
	}
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn, err := redis.Dial("tcp", "6379")
	if err != nil {
		resp["code"] = 500
		resp["errmsg"] = "redis连接错误"
		this.Data["json"] = resp
		return
	}
	defer conn.Close()

	conn.Do("hset", "cart_"+strconv.Itoa(user.Id), skuid, count)

	resp["code"] = 200
	resp["msg"] = "ok"
	this.Data["json"] = resp

}

// 删除购物车数据
func (this *CartController) HandleDeleteCart() {

	skuid, err := this.GetInt("skuid")

	resp := make(map[string]interface{})

	if err != nil {
		resp["code"] = 400
		resp["errmsg"] = "请求数据不正确"

		this.Data["json"] = resp
	}
	defer this.ServeJSON()

	conn, err := redis.Dial("tcp", "6379")
	if err != nil {
		resp["code"] = 500
		resp["errmsg"] = "redis连接错误"
		this.Data["json"] = resp
		return
	}
	defer conn.Close()

	userName := this.GetSession("userName")
	o := orm.NewOrm()
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), skuid)

	// 返回数据
	resp["code"] = 200
	resp["msg"] = "ok"
	this.Data["json"] = resp
}
