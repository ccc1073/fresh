package controllers

import (
	"encoding/base64"
	"fresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/utils"
	"github.com/gomodule/redigo/redis"
	"regexp"
	"strconv"
)

type UserController struct {
	beego.Controller
}

// 展示注册页面
func (this *UserController) ShowReg() {

	this.TplName = "register.html"
}

// 处理注册数据
func (this *UserController) HandleReg() {

	userName := this.GetString("user_name")
	pwd := this.GetString("pwd")
	cpwd := this.GetString("cpwd")
	email := this.GetString("email")

	// 校验数据
	if userName == "" || pwd == "" || cpwd == "" || email == "" {
		this.Data["errmsg"] = "数据不完整请重新注册"
		this.TplName = "register.html"
		return
	}

	if pwd != cpwd {
		this.Data["errmsg"] = "两次密码输入不一致 请重新注册"
		this.TplName = "register.html"
		return
	}

	reg, _ := regexp.Compile("^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\\.[a-zA-Z0-9_-]+)+$")
	res := reg.FindString(email)

	if res == "" {
		this.Data["errmsg"] = "邮箱格式不正确"
		this.TplName = "register.html"
		return
	}
	// 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	user.PassWord = pwd
	user.Email = email

	_, err := o.Insert(&user)
	if err != nil {
		this.Data["errmsg"] = "注册失败"
		this.TplName = "register.html"
		return
	}
	// 发送邮件
	emailConifg := `{"username":"xxx@qq.com","password":"xxxxx","host":"smtp.qq.com","port":"587"}`
	emailConn := utils.NewEMail(emailConifg)
	emailConn.From = "天天生鲜系统注册服务"
	emailConn.To = []string{email}
	emailConn.Subject = "天天生鲜用户注册"
	// 这里发送给用户的时激活请求地址
	emailConn.Text = "www.baidu.com/active?id=" + strconv.Itoa(user.Id)

	emailConn.Send()

	// 返回视图d
	this.Ctx.WriteString("注册成功,请去邮箱进行激活")
}

// 激活处理
func (this *UserController) ActiveUser() {

	id, err := this.GetInt("id")

	if err != nil {
		this.Data["errmsg"] = "激活的用户不存在"
		this.TplName = "register.html"
	}
	// 处理数据
	o := orm.NewOrm()
	var user models.User
	user.Id = id

	err = o.Read(&user)
	if err != nil {
		this.Data["errmag"] = "用户不存在"
		this.TplName = "register.html"
		return
	}

	user.Active = true
	o.Update(&user)
	// 返回视图
	this.Redirect("/login", 302)
}

// 展示登录页面
func (this *UserController) ShowLogin() {
	userName := this.Ctx.GetCookie("userName")

	temp, _ := base64.StdEncoding.DecodeString(userName)

	if string(temp) == "" {
		this.Data["errmsg"] = ""
		this.Data["checked"] = ""
	} else {
		this.Data["userName"] = string(temp)
		this.Data["checked"] = "checked"
	}

	this.TplName = "login.html"
}

// 用户登录
func (this *UserController) Handlelogin() {
	userName := this.GetString("username")
	pwd := this.GetString("pwd")

	//校验数据
	if userName == "" || pwd == "" {
		this.Data["errmsg"] = "登录数据不能为空"
		this.TplName = "login.html"
		return
	}

	// 处理数据
	o := orm.NewOrm()
	var user models.User

	user.Name = userName

	err := o.Read(&userName)

	if err != nil {
		this.Data["errmsg"] = "用户名或密码错误"
		this.TplName = "login.html"
		return
	}

	if user.PassWord != pwd {
		this.Data["errmsg"] = "用户名或密码错误"
		this.TplName = "login.html"
		return
	}

	if user.Active != true {
		this.Data["errmsg"] = "用户未激活"
		this.TplName = "login.html"
		return
	}
	// base64加密
	remember := this.GetString("remember")
	if remember == "on" {
		tmp := base64.StdEncoding.EncodeToString([]byte(userName))

		this.Ctx.SetCookie("userName", tmp, 24*3600)
	} else {
		this.Ctx.SetCookie("userName", userName, -1)

	}
	this.SetSession("userName", userName)
	//this.Ctx.WriteString("登录成功")
	this.Redirect("/", 302)

}

// 用户退出
func (this *UserController) Logout() {
	this.DelSession("userName")
	// 跳转视图
	this.Redirect("/login", 302)
}

// 展示用户中心信息
func (this *UserController) ShowUserCenterInfo() {

	userName := GetUser(&this.Controller)
	this.Data["userName"] = userName

	o := orm.NewOrm()
	var addr models.Address
	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName).Filter("Isdefault", true).One(&addr)

	if addr.Id == 0 {
		this.Data["addr"] = ""
	} else {
		this.Data["addr"] = addr
	}
	// 获取历史浏览记录
	conn, err := redis.Dial("tcp", "6379")

	defer conn.Close()
	if err != nil {
		beego.Info("连接错误")
	}
	// 获取用户id
	var user models.User
	user.Name = userName
	rep, _ := conn.Do("lrange", "history_"+strconv.Itoa(user.Id), 0.4)

	goodsIds, _ := redis.Ints(rep, err)

	var goodsSkus []models.GoodsSKU

	for _, value := range goodsIds {

		var goods models.GoodsSKU
		goods.Id = value
		o.Read(&goods)
		goodsSkus = append(goodsSkus, goods)
	}

	beego.Info(goodsSkus)
	this.Data["goodsSKUs"] = goodsSkus

	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_info.html"
}

// 展示用户中心订单
func (this *UserController) ShowUserCenterOrder() {

	userName := GetUser(&this.Controller)

	o := orm.NewOrm()
	var user models.User
	user.Name = userName
	o.Read(&user, "Name")

	// 获取订单表的数据
	var orderInfos []models.OrderInfo
	o.QueryTable("OrderInfo").RelatedSel("User").Filter("User__Id", user.Id).All(&orderInfos)

	goodsBuffer := make([]map[string]interface{}, len(orderInfos))

	for index, orderInfo := range orderInfos {

		var orderGoods models.OrderGoods
		o.QueryTable("OrderGoods").RelatedSel("OrderInfo", "GoodsSKU").Filter("OrderInfo_Id", orderInfo.Id).All(&orderGoods)

		temp := make(map[string]interface{})

		temp["orderInfo"] = orderInfos
		temp["orderGoods"] = orderGoods

		goodsBuffer[index] = temp

	}

	this.Data["goodsBuffer"] = goodsBuffer

	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_order.html"

}

// 展示用户中心地址页
func (this *UserController) ShowUserCenterSite() {

	userName := GetUser(&this.Controller)
	//this.Data["userName"] = userName
	// 获取地址信息
	o := orm.NewOrm()
	var addr models.Address

	o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName).Filter("Isdefault", true).One(&addr)

	this.Data["addr"] = addr

	this.Layout = "userCenterLayout.html"
	this.TplName = "user_center_site.html"
}

// 添加收货地址
func (this *UserController) HandleCenterSite() {

	receiver := this.GetString("receiver")
	addr := this.GetString("addr")
	zipcode := this.GetString("zipcode")
	phone := this.GetString("phone")

	if receiver == "" || addr == "" || zipcode == "" || phone == "" {
		this.Redirect("/user/userCenterSite", 302)
		return
	}

	o := orm.NewOrm()
	userName := this.GetSession("userName")
	var user models.User
	user.Name = userName.(string)
	o.Read(&user, "Name")

	var addrUser models.Address

	//addrUser.Isdefault = true
	err := o.QueryTable("Address").RelatedSel("User").Filter("User__Name", userName).Filter("Isdefault", true).One(&addrUser)
	if err == nil {
		addrUser.Isdefault = false
		o.Update(&addrUser)
	}

	var addrUserNew models.Address

	addrUserNew.Addr = addr
	addrUserNew.Receiver = receiver
	addrUserNew.Zipcode = zipcode
	addrUserNew.Phone = phone
	addrUserNew.User = &user
	o.Insert(&addrUser)

	this.Redirect("/user/userCenterSite", 302)
}
