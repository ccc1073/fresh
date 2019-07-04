package routers

import (
	"fresh/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {
	beego.InsertFilter("/user/*", beego.BeforeExec, filterFunc)

	//beego.Router("/", &controllers.MainController{})
	//beego.Router("/register",&controllers.UserController{},"get:ShowReg;post:HandleReg")
	// 激活用户
	beego.Router("/active", &controllers.UserController{}, "get:ActiveUser")
	// 用户登录
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:Handlelogin")
	// 跳转首页
	beego.Router("/", &controllers.GoodsController{}, "get:ShowIndex")
	// 用户退出
	beego.Router("/user/logout", &controllers.UserController{}, "get:Logout")
	// 用户中心
	beego.Router("/user/usercenterinfo", &controllers.UserController{}, "get:ShowUserCenterInfo")
	// 用户中心订单页
	beego.Router("/user/usercenterorder", &controllers.UserController{}, "get:ShowUserCenterOrder")
	// 用户中心地址页
	beego.Router("/user/usercentersite", &controllers.UserController{}, "get:ShowUserCenterSite;post:HandleCenterSite")
	// 商品详情展示
	beego.Router("/goodsDetail", &controllers.GoodsController{}, "get:ShowGoodsDetail")
	// 展示商品列表页
	beego.Router("/goodsList", &controllers.GoodsController{}, "get:ShowList")
	// 商品搜索
	beego.Router("/goodsSearch", &controllers.GoodsController{}, "post:HandleSearch")
	// 添加购物车
	beego.Router("/user/addCart", &controllers.CartController{}, "post:HandleAddcart")
	// 展示购物车
	beego.Router("/user/cart", &controllers.CartController{}, "get:ShowCart")
	// 添加购物车数量
	beego.Router("/user/updateCart", &controllers.CartController{}, "post:HandleUpdateCart")
	// 删除购物车
	beego.Router("/user/deletecart", &controllers.CartController{}, "post:HandleDeleteCart")
	// 展示订单页面
	beego.Router("user/showOrder", &controllers.OrderController{}, "get:showOrder")
	// 提交订单
	beego.Router("user/addOrder", &controllers.OrderController{}, "post:HandleaddOrder")
	// 支付
	beego.Router("/user/pay", &controllers.OrderController{}, "get:HandlePay")

	//支付成功
	beego.Router("/user/payok", &controllers.OrderController{}, "get:PayOk")

}

// 过滤器
var filterFunc = func(ctx *context.Context) {

	userName := ctx.Input.Session("userName")

	if userName == nil {
		ctx.Redirect(302, "/login")
		return
	}

}
