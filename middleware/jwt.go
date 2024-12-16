package middleware

import (
	"baby/models"
	"baby/settings"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

type CustomClaims struct {
	Username string `json:"username"`
	UserId   int64  `json:"userId"`
	jwt.RegisteredClaims
}

// 使用GenToken生成JWT
func GenToken(username string, userId int64) (string, error) {
	expire := time.Now().Add(settings.TokenExpireDuration)
	claims := CustomClaims{
		username,
		userId,
		jwt.RegisteredClaims{
			Issuer: "Autsdd",
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := t.SignedString(settings.Secret)
	j := models.Jwts{Token: token, Expire: expire}
	models.DB.Create(&j)
	return token, err
}

// ParseToken解析JWT
func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return settings.Secret, nil
		})
	if err != nil {
		return nil, err
	}
	//校验token
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func JWTAuthMiddleware(c *gin.Context) {
	//token放在请求头中
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg":   "请求头的Authorization为空",
		})
		c.Abort()
		return
	}
	mc, err := ParseToken(authHeader)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg":   "无效的Token",
		})
		c.Abort()
		return
	}
	var jwts models.Jwts
	models.DB.Where("token = ?", authHeader).First(&jwts)
	if jwts.Token != "" {
		if jwts.Expire.After(time.Now()) {
			jwts.Expire = time.Now().Add(settings.TokenExpireDuration)
			models.DB.Save(&jwts)
		} else {
			models.DB.Unscoped().Delete(&jwts)
		}
	} else {
		c.JSON(http.StatusOK, gin.H{
			"state": "fail",
			"msg":   "无效的Token",
		})
		c.Abort()
		return
	}
	c.Set("username", mc.Username)
	c.Set("userId", mc.UserId)
	c.Next()
}
