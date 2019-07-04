package controllers

import (
	"fresh/models"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"strconv"
)

type GoodsController struct {
	beego.Controller
}

func GetUser(this *beego.Controller) string {
	userName := this.GetSession("userName")

	if userName == "" {
		this.Data["userName"] = ""
	} else {
		this.Data["userName"] = userName.(string)
		return userName.(string)
	}
	return ""
}

func ShowLaout(this *beego.Controller) {
	// 查询类型
	o := orm.NewOrm()
	var types []models.GoodsType

	o.QueryTable("GoodsType").All(&types)

	this.Data["types"] = types
	GetUser(this)
	this.Layout = "goodsLayout.html"
}

func PageTool(pageCount int, pageIndex int) []int {

	var pages []int

	if pageCount <= 5 {
		pages = make([]int, pageCount)
		for i, _ := range pages {
			pages[i] = i + 1
		}
	} else if pageIndex <= 3 {
		pages = []int{1, 2, 3, 4, 5}
	} else if pageIndex > pageCount-3 {
		pages = []int{pageCount - 4, pageCount - 3, pageCount - 2, pageCount - 1, pageCount}
	} else {
		pages = []int{pageIndex - 2, pageIndex - 1, pageIndex, pageIndex + 1, pageIndex + 2}
	}

	return pages
}

// 展示首页
func (this *GoodsController) ShowIndex() {

	GetUser(&this.Controller)
	// 获取类型数据
	o := orm.NewOrm()
	var goodsTypes []models.GoodsType
	o.QueryTable("GoodsType").All(&goodsTypes)

	this.Data["goodsTypes"] = goodsTypes
	// 获取轮播图数据
	var indexGoodsBanner []models.IndexGoodsBanner
	o.QueryTable("IndexGoodsBanner").OrderBy("Index").All(&indexGoodsBanner)
	this.Data["indexGoodsBanner"] = indexGoodsBanner

	// 获取促销商品数据
	var promotionGoods []models.IndexPromotionBanner
	o.QueryTable("IndexPromotionBanner").OrderBy("Index").All(&promotionGoods)
	this.Data["promotionsGoods"] = promotionGoods

	// 首页展示商品数据
	goods := make([]map[string]interface{}, len(goodsTypes))

	//向切片interface中插入类型数据
	for index, value := range goodsTypes {
		//获取对应类型的首页展示商品
		temp := make(map[string]interface{})
		temp["type"] = value
		goods[index] = temp
	}
	//商品数据

	for _, value := range goods {
		var textGoods []models.IndexTypeGoodsBanner
		var imgGoods []models.IndexTypeGoodsBanner
		//获取文字商品数据
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 0).All(&textGoods)
		//获取图片商品数据
		o.QueryTable("IndexTypeGoodsBanner").RelatedSel("GoodsType", "GoodsSKU").OrderBy("Index").Filter("GoodsType", value["type"]).Filter("DisplayType", 1).All(&imgGoods)

		value["textGoods"] = textGoods
		value["imgGoods"] = imgGoods
	}
	this.Data["goods"] = goods

	this.TplName = "index.html"
}

// 展示商品详情页面
func (this *GoodsController) ShowGoodsDetail() {

	id, err := this.GetInt("id")

	if err != nil {
		beego.Error("数据错误")
		this.Redirect("/", 302)
		return
	}

	o := orm.NewOrm()

	var goodsSku models.GoodsSKU
	goodsSku.Id = id
	o.Read(&goodsSku)
	// 返回视图
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType", "Goods").Filter("Id", id).One(&goodsSku)

	var goodsNew []models.GoodsSKU

	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType", goodsSku.GoodsType).OrderBy("Time").Limit(2, 0).All(&goodsNew)

	this.Data["goodsNew"] = goodsNew

	this.Data["goodsSku"] = goodsSku

	// 添加历史浏览记录
	// 判断用户是否登录
	userName := this.GetSession("userName")
	if userName != nil {
		o := orm.NewOrm()
		var user models.User
		user.Name = userName.(string)
		o.Read(&user, "Name")
		// 添加历史记录 用redis存储
		conn, err := redis.Dial("tcp", "6379")

		if err != nil {
			beego.Info("redis连接错误")

		}
		defer conn.Close()
		// 将以前苏香桐商品的历史浏览记录删除
		conn.Do("lrem", "history_"+strconv.Itoa(user.Id), 0, id)

		conn.Do("lpush", "history_"+strconv.Itoa(user.Id), id)

	}

	ShowLaout(&this.Controller)
	this.TplName = "detail.html"

}

// 展示商品列表页
func (this *GoodsController) ShowList() {

	id, err := this.GetInt("typeId")

	if err != nil {
		beego.Info("请求路径错误")
		this.Redirect("/", 302)
		return
	}

	ShowLaout(&this.Controller)
	// 获取新品
	o := orm.NewOrm()
	var goodsNew []models.GoodsSKU
	o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).OrderBy("Time").Limit(2, 0).All(&goodsNew)

	this.Data["goodsNew"] = goodsNew

	// 分页实现
	count, _ := o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodType__Id", id).Count()
	pageSzie := 1
	pageCount := math.Ceil(float64(count) / float64(pageSzie))

	pageIndex, err := this.GetInt("pageIndex")
	if err != nil {
		pageIndex = 1
	}
	pages := PageTool(int(pageCount), pageIndex)

	this.Data["pages"] = pages
	this.Data["typeId"] = id
	this.Data["pageIndex"] = pageIndex

	var goods []models.GoodsSKU
	start := (pageIndex - 1) * pageSzie

	// 获取上一页页码
	prePage := pageIndex - 1
	if prePage <= 1 {
		prePage = 1
	}
	this.Data["prePage"] = prePage
	//获取下一页页码
	nextPage := pageIndex + 1
	if nextPage > int(pageCount) {
		nextPage = int(pageCount)
	}
	this.Data["nextPage"] = nextPage

	// 按照一定顺序获取商品
	sort := this.GetString("sort")
	if sort == "" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).Limit(pageSzie, start).All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = ""

	} else if sort == "price" {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).Limit(pageSzie, start).OrderBy("Price").All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = "price"

	} else {
		o.QueryTable("GoodsSKU").RelatedSel("GoodsType").Filter("GoodsType__Id", id).Limit(pageSzie, start).OrderBy("Sale").All(&goods)
		this.Data["goods"] = goods
		this.Data["sort"] = "sale"
	}

	//返回视图
	this.TplName = "list.html"

}

// 处理搜索
func (this *GoodsController) HandleSearch() {

	goodsName := this.GetString("goodsName")
	o := orm.NewOrm()
	var goods []models.GoodsSKU
	if goodsName == "" {
		o.QueryTable("GoodsSKU").All(&goods)
		this.Data["goods"] = goods

		ShowLaout(&this.Controller)
		this.TplName = "search.html"

	}
	// 处理数据
	o.QueryTable("GoodsSKU").Filter("Name__icontains", goodsName).All(&goods)

	this.Data["goods"] = goods
	ShowLaout(&this.Controller)
	this.TplName = "search.html"
}
