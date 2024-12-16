package v1

import (
	"baby/middleware"
	"baby/models"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func Home(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	//今日必抢商品信息
	var commodity []models.Commodities
	models.DB.Order("created_at DESC").Find(&commodity)
	data["commodityInfos"] = [][]models.Commodities{
		commodity[0:4], commodity[4:8]}
	//分类商品信息
	var classification = map[string]string{"clothes": "儿童服饰",
		"food": "奶粉辅食", "goods": "儿童用品"}
	for key, value := range classification {
		var types = []string{}
		var temp []models.Commodities
		models.DB.Model(&models.Types{}).Where("firsts = ?", value).Select("seconds").Find(&types)
		models.DB.Where("types in ?", types).Order("created_at DESC").Find(&temp)
		data[key] = temp[0:5]
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

// 商品列表
func CommodityList(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	//获取请求参数
	types := c.DefaultQuery("types", "")
	search := c.DefaultQuery("search", "")
	sort := c.DefaultQuery("sort", "")
	page := c.DefaultQuery("page", "1")
	p, _ := strconv.Atoi(page)
	//商品分类列表
	var firsts = []string{}
	models.DB.Model(&models.Types{}).Distinct("firsts").Find(&firsts)
	var res []map[string]interface{}
	for _, first := range firsts {
		var seconds = []string{}
		models.DB.Model(&models.Types{}).Where("firsts = ?", first).
			Select("seconds").Find(&seconds)
		res = append(res, map[string]interface{}{"name": first, "value": seconds})
	}
	data["types"] = res
	//商品列表信息
	var commodity []models.Commodities
	querys := models.DB.Model(&models.Commodities{})
	if types != "" {
		querys = querys.Where("types in ?", types)
	}
	if sort != "" {
		querys = querys.Order(sort + "DESC")
	}
	if search != "" {
		querys = querys.Where("name like ?", "%"+search+"%")
	}
	//分页
	querys, previous, next, count, pageCount := models.Paginate(querys, p)
	querys = querys.Find(&commodity)
	data["commodityInfos"] = map[string]interface{}{"data": commodity, "previous": previous, "next": next,
		"count": count, "pageCount": pageCount}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

// CommodityDetail 商品详情
func CommodityDetail(c *gin.Context) {
	context := gin.H{"state": "success", "msg": "获取成功"}
	data := gin.H{}
	id := c.Param("id")
	//获取商品详细信息
	var commodity models.Commodities
	models.DB.Where("id = ?", id).First(&commodity)
	data["commodities"] = commodity
	//获取推荐商品
	var recommend []models.Commodities
	models.DB.Where("id != ?", id).Order("sold DESC").Limit(5).Find(&recommend)
	data["recommend"] = recommend
	//收藏状态
	data["likes"] = false
	//获取请求头的Authorization
	//获取用户信息，根据用户信息和商品ID查找收藏记录
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader != "" {
		mc, _ := middleware.ParseToken(authHeader)
		if mc != nil {
			UserId := mc.UserId
			if UserId != 0 {
				var records []models.Records
				models.DB.Where("user_id = ? and commodity_id = ?", UserId, id).Find(&records)
				if len(records) > 0 {
					data["likes"] = true
				}
			}
		}
	}
	context["data"] = data
	c.JSON(http.StatusOK, context)
}

// 收藏
func CommodityCollect(c *gin.Context) {
	context := gin.H{"state": "fail", "msg": "收藏失败"}
	data, _ := c.GetRawData()
	var body map[string]int64
	json.Unmarshal(data, &body)
	id := body["id"]
	userId, _ := c.Get("userId")
	fmt.Println(userId)
	var records []models.Records
	models.DB.Where("user_id = ? and commodity_id = ?", userId.(int64), id).Find(&records)
	if len(records) == 0 {
		models.DB.Model(&models.Commodities{}).Where("id = ?", id).Update("likes", 1)
		r := models.Records{UserId: userId.(int64), CommodityId: id}
		models.DB.Create(&r)
		context["msg"] = "收藏成功"
		context["state"] = "success"
	} else {
		context["msg"] = "收藏取消"
		context["state"] = "success"
		models.DB.Unscoped().Delete(&records)
	}
	c.JSON(http.StatusOK, context)
}
