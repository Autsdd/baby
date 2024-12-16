package settings

import (
	"github.com/gin-gonic/gin"
	"time"
)

type Database struct {
	User     string
	Password string
	Host     string
	Name     string
}

var MySQLSetting = &Database{
	User:     "root",
	Password: "123456",
	Host:     "127.0.0.1:3306",
	Name:     "baby",
}

var Mode = gin.ReleaseMode

// token有效期
var TokenExpireDuration = time.Minute * 30

var Secret = []byte("你好")

// 分页，一页6条数据
var PageSize = 6

// 支付宝沙箱信息
var AppId = ""
var AlipayPublicKeyString = ""
var AppPrivateKeyString = ""
