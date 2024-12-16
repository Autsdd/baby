package v1

import (
	"baby/middleware"
	"baby/models"
	"baby/settings"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ShopperLogin(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "注册或登录失败"}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	c.BindJSON(&body)
	username := body.Username
	p := body.Password
	if username != "" && p != "" {
		context["state"] = "success"
		//生成登录时间
		lastLogin := time.Now()
		context["last_login"] = lastLogin.Format("2006-01-02 15:04:05")
		//密码加密，用于查找用户数据
		m := md5.New()
		m.Write([]byte(p))
		password := hex.EncodeToString(m.Sum(nil))
		//查找用户，用户存在则登录成功，不存在则创建
		var userID uint
		var users models.Users
		models.DB.Where("username = ?", username).First(&users)
		if users.ID > 0 {
			if users.Password == password {
				userID = users.ID
				users.LastLogin = lastLogin
				models.DB.Save(&users)
				context["msg"] = "登录成功"
			} else {
				context["msg"] = "请输入正确密码"
				context["state"] = "fail"
			}
		} else {
			context["msg"] = "注册成功"
			r := models.Users{Username: username, Password: p,
				IsStaff: 1, LastLogin: lastLogin}
			models.DB.Create(&r)
			if r.ID > 0 {
				userID = r.ID
			} else {
				context["msg"] = "注册失败"
				context["state"] = "fail"
			}
		}
		//创建token
		token := ""
		if userID > 0 {
			token, _ = middleware.GenToken(username, int64(userID))
		}
		context["token"] = token
	}
	c.JSON(http.StatusOK, context)
}

func ShopperLogout(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "退出失败"}
	userId, _ := c.Get("user_id")
	if userId == nil {
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader != "" {
			var jwts models.Jwts
			models.DB.Where("token = ?", authHeader).First(&jwts)
			models.DB.Unscoped().Delete(&jwts)
			context = gin.H{"state": "success", "msg": "退出成功"}
		}
	}
	c.JSON(http.StatusOK, context)
}

func ShopperShopCart(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	userId, _ := c.Get("userId")
	if c.Request.Method == "GET" {
		if userId != 0 {
			var carts []models.Carts
			models.DB.Preload("Commodities").Where("user_id = ?", userId.(int64)).Order("id DESC").Find(&carts)
			context["data"] = carts
		}
	}
	if c.Request.Method == "POST" {
		context = gin.H{"state": "fail", "msg": "加购失败"}
		var body struct {
			Id       int `json:"id"`
			Quantity int `json:"quantity"`
		}
		c.BindJSON(&body)
		id := body.Id
		quantity := body.Quantity
		var commodity models.Commodities
		models.DB.Where("id = ?", id).First(&commodity)
		//查找商品是否存在
		if commodity.ID > 0 {
			//购物车同一个商品，只增加购买数量
			var cart models.Carts
			models.DB.Where("commodity_id = ? and"+
				"user_id = ?", id, userId.(int64)).First(&cart)
			if cart.ID > 0 {
				cart.Quantity += quantity
				models.DB.Save(&cart)
			} else {
				carts := models.Carts{UserId: int(userId.(int64)),
					CommodityId: strconv.Itoa(id), Quantity: quantity}
				models.DB.Create(&carts)
			}
			context = gin.H{"state": "success", "msg": "加购成功"}
		}
	}
	c.JSON(http.StatusOK, context)
}

// 购物车功能
func ShopperDelete(c *gin.Context) {
	var body struct {
		CartId int64 `json:"cartId"`
	}
	c.BindJSON(&body)
	cartId := body.CartId
	var cart []models.Carts
	if cartId != 0 {
		models.DB.Where("id = ?", cartId).First(&cart)
	} else {
		userId, _ := c.Get("userId")
		models.DB.Where("user_id = ?", userId.(int64)).First(&cart)
	}
	models.DB.Unscoped().Delete(&cart)
	context := gin.H{"state": "success", "msg": "删除成功"}
	c.JSON(http.StatusOK, context)
}

// 支付宝apy
func ShopperPays(c *gin.Context) {
	var body struct {
		Total   string   `json:"total"`
		PayInfo string   `json:"payInfo"`
		CartID  []string `json:"cartID"`
	}
	c.BindJSON(&body)
	total := strings.Replace(body.Total, "￥", "", -1)
	payInfo := body.PayInfo
	cartId := body.CartID
	//根据请求参数的payInfo获取订单信息
	//如果不在，则从购物车创建订单
	//如果存在，则从个人中心继续支付之前的订单
	if total == "" {
		context := gin.H{"state": "fail",
			"msg": "支付失败,请输入金额"}
		c.JSON(http.StatusOK, context)
	}
	if payInfo == "" {
		payInfo = strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	userId, _ := c.Get("userId")
	var order models.Orders
	models.DB.Where("pay_info = ?", payInfo).First(&order)
	if order.ID == 0 {
		carts := models.Orders{UserId: userId.(int64), Price: total, PayInfo: payInfo, State: 0}
		models.DB.Create(&carts)
	}
	//删除购物车对应信息
	if len(cartId) > 0 {
		models.DB.Unscoped().Delete(&[]models.Carts{}, cartId)
	}
	//调用支付宝接口
	client, _ := alipay.New(settings.AppId, settings.AppPrivateKeyString, false)
	client.LoadAlipayCertPublicKey(settings.AlipayPublicKeyString)
	var p = alipay.TradePagePay{}
	p.ReturnURL = "http://localhost:8010/#/shopper"
	p.Body = "支付宝测试"
	p.Subject = "测试"
	p.OutTradeNo = payInfo
	p.TotalAmount = total
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url, _ := client.TradePagePay(p)
	payURL := url.String()
	//响应内容
	context := gin.H{"state": "success", "msg": "支付成功", "data": payURL}
	fmt.Println(payURL)
	c.JSON(http.StatusOK, context)
}

// 个人中心
func ShopperHome(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	userId, _ := c.Get("userId")
	payInfo := c.DefaultQuery("out_trade_no", "")
	if payInfo != "" {
		models.DB.Model(&models.Orders{}).Where("pay_info = ?", payInfo).Update("state", 1)
	}
	if userId != 0 {
		var orders []models.Orders
		models.DB.Where("user_id = ?", userId).Order("id DESC").Find(&orders)
		data["orders"] = orders
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}
