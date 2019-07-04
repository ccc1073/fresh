package controllers

import (
	"fmt"
	"fresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"github.com/smartwalle/alipay"
	"strconv"
	"strings"
	"time"
)

type OrderController struct {
	beego.Controller
}

// 展示支付订单
func (this *OrderController) showOrder() {

	skuids := this.GetStrings("skuid")

	// 校验数据
	if len(skuids) == 0 {
		beego.Info("请求数据错误")
		this.Redirect("/user/cart", 302)
		return
	}

	// 处理数据
	o := orm.NewOrm()

	// 获取用户数据
	var user models.User
	userName := this.GetSession("userName")
	user.Name = userName.(string)
	o.Read(&user, "Name")

	conn, _ := redis.Dial("tcp", "6379")

	defer conn.Close()
	goodsBuffer := make([]map[string]interface{}, len(skuids))
	totalPrice := 0
	totalCount := 0

	for index, skuid := range skuids {
		temp := make(map[string]interface{})

		id, _ := strconv.Atoi(skuid)

		var goodsSku models.GoodsSKU

		goodsSku.Id = id
		o.Read(&goodsSku)

		temp["goods"] = goodsSku
		// 获取商品数量
		resp, err := conn.Do("hget", "cart_"+strconv.Itoa(user.Id), id)

		count, _ := redis.Int(resp, err)
		temp["count"] = count

		// 小计
		amount := goodsSku.Price * count
		temp["amount"] = amount

		goodsBuffer[index] = temp

		// 计算总金额和总件数
		totalCount += count
		totalPrice += amount

	}
	this.Data["userName"] = userName

	this.Data["goodsBuffer"] = goodsBuffer
	// 获取地址数据
	var addr []models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Id", user.Id).All(&addr)

	this.Data["addrs"] = addr

	this.Data["totalCount"] = totalCount

	this.Data["totalPrice"] = totalPrice
	// 运费
	tranferPrice := 10
	this.Data["transferPrice"] = tranferPrice

	// 总计
	this.Data["AllPrice"] = totalPrice + tranferPrice

	// 传递所有商品的id
	this.Data["skuids"] = skuids

	this.TplName = "place_order.html"
}

// 提交订单
func (this *OrderController) HandleaddOrder() {

	addrid, _ := this.GetInt("addrid")
	payId, _ := this.GetInt("payId")
	skuIds := this.GetString("skuids")
	//totalPrice,_ := this.GetInt("totalPrice")
	totalCount, _ := this.GetInt("totalCount")
	transferPrice, _ := this.GetInt("transferPrice")
	AllPrice, _ := this.GetInt("AllPrice")

	ids := skuIds[1 : len(skuIds)-1]
	skuids := strings.Split(ids, " ")

	if len(skuIds) == 0 {

		return
	}
	userName := this.GetSession("userName")

	o := orm.NewOrm()

	// 开启事务
	o.Begin()

	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "userName")

	var order models.OrderInfo
	order.OrderId = time.Now().Format("2006010215030405") + strconv.Itoa(user.Id)
	order.User = &user
	order.Orderstatus = 1
	order.PayMethod = payId
	order.TotalCount = totalCount
	order.TotalPrice = AllPrice
	order.TransitPrice = transferPrice

	resp := make(map[string]interface{})

	defer this.ServeJSON()

	// 查询地址
	var addr models.Address
	addr.Id = addrid
	o.Read(&addr)

	order.Address = &addr

	o.Insert(&order)

	conn, _ := redis.Dial("tcp", "6379")
	defer conn.Close()
	// 向订单商品表中插入数据

	for _, skuid := range skuids {
		id, _ := strconv.Atoi(skuid)
		var goods models.GoodsSKU
		goods.Id = id
		o.Read(&goods)

		var orderGoods models.OrderGoods
		orderGoods.OrderInfo = &order

		count, _ := redis.Int(conn.Do("hget", "cart_"+strconv.Itoa(user.Id), id))

		if count > goods.Stock {
			resp["code"] = 400
			resp["errmsg"] = "商品库存不足"
			this.Data["json"] = resp
			o.Rollback()
			return
		}
		preCount := goods.Stock

		orderGoods.Count = count
		orderGoods.Price = goods.Price * count

		o.Insert(&orderGoods)

		goods.Stock -= count
		goods.Sales += count

		o.Update(&goods)
		// update参数中为更新的字段
		updateCount, _ := o.QueryTable("GoodSKU").Filter("Id", goods.Id).Filter("Stock", preCount).Update(orm.Params{"Stock": goods.Stock, "Sales": goods.Sales})

		if updateCount == 0 {

			resp["code"] = 400
			resp["errmsg"] = "更新失败"
			this.Data["json"] = resp
			o.Rollback()
			return
		}

		conn.Do("hdel", "cart_"+strconv.Itoa(user.Id), goods.Id)

		// 修改mysql事物级别打开mysqld.cnf文件 transaction-isolation = READ-COMMITTED

	}
	o.Commit()
	resp["code"] = 200
	resp["msg"] = "ok"
	resp["json"] = resp

}

// 处理支付
func (this *OrderController) HandlePay() {

	var aliPublicKey = ""
	var priveteKey = "xx"
	var appId = "xxx"
	client, _ := alipay.New(appId, aliPublicKey, priveteKey, false)

	orderId := this.GetString("orderId")
	totalPrice := this.GetString("totalPrice")
	fmt.Println(orderId, totalPrice)

	var p = alipay.TradeWapPay{}
	p.NotifyURL = "xxxx"
	p.ReturnURL = "xxx"
	p.Subject = "标题"
	p.OutTradeNo = "唯一单号"
	p.TotalAmount = "10.00"
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	var url, err = client.TradeWapPay(p)

	if err != nil {
		return
	}
	var payUrl = url.String()

	//this.Data["payUrl"] = payUrl
	this.Redirect(payUrl, 302)
}

//支付成功
func (this *OrderController) PayOk() {
	//获取数据
	//out_trade_no=999998888777
	orderId := this.GetString("out_trade_no")
	this.GetControllerAndAction()

	//校验数据
	if orderId == "" {
		beego.Info("支付返回数据错误")
		this.Redirect("/user/userCenterOrder", 302)
		return
	}

	//操作数据

	o := orm.NewOrm()
	count, _ := o.QueryTable("OrderInfo").Filter("OrderId", orderId).Update(orm.Params{"Orderstatus": 2})
	if count == 0 {
		beego.Info("更新数据失败")
		this.Redirect("/user/userCenterOrder", 302)
		return
	}

	//返回视图
	this.Redirect("/user/userCenterOrder", 302)
}
