package routers

import (
	"baby/middleware"
	v1 "baby/servers/v1"
	"baby/settings"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func InitRoute() *gin.Engine {
	gin.SetMode(settings.Mode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.StaticFS("/static", http.Dir("static"))
	//配置跨域访问
	config := cors.DefaultConfig()
	//允许所有域名
	config.AllowAllOrigins = true
	//允许执行的请求方法
	config.AllowMethods = []string{"GET", "POST"}
	//允许执行的请求头
	config.AllowHeaders = []string{"tus-resumable", "upload-length",
		"upload-metadata", "cache-control",
		"x-requested-with", "*"}
	r.Use(cors.New(config))
	//定义路由
	apiv1 := r.Group("/api/v1")
	commodity := apiv1.Group("")
	{
		//网站首页
		commodity.GET("home/", v1.Home)
		//商品列表
		commodity.GET("commodity/list/", v1.CommodityList)
		//商品详情
		commodity.GET("commodity/detail/:id/", v1.CommodityDetail)
		//用户注册登陆
		commodity.POST("shopper/login/", v1.ShopperLogin)
	}
	shopper := apiv1.Group("", middleware.JWTAuthMiddleware)
	{
		//商品收藏
		shopper.POST("commodity/collect/", v1.CommodityCollect)
		//退出登陆
		shopper.POST("shopper/logout/", v1.ShopperLogout)
		//个人主页
		shopper.GET("shopper/home", v1.ShopperHome)
		//加入购物车
		shopper.GET("shopper/shopcart/", v1.ShopperShopCart)
		//加入购物车
		shopper.POST("shopper/shopcart/", v1.ShopperShopCart)
		//在线支付
		shopper.POST("shopper/pays/", v1.ShopperPays)
		//删除购物车商品
		shopper.POST("shopper/delete", v1.ShopperDelete)
	}
	return r
}
