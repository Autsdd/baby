package models

import (
	"baby/settings"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

// 数据表结构体
type Types struct {
	gorm.Model
	Firsts  string `json:"firsts" gorm:"type:varchar(255)"`
	Seconds string `json:"seconds" gorm:"type:varchar(255)"`
}

type Commodities struct {
	gorm.Model
	Name     string    `json:"name" gorm:"type:varchar(255)"`
	Sizes    string    `json:"size" gorm:"type:varchar(255)"`
	Types    string    `json:"types" gorm:"type:varchar(255)"`
	Price    float32   `json:"price"`
	Discount float32   `json:"discount"`
	Stock    int       `json:"stock"`
	Likes    int       `json:"likes"`
	Created  time.Time `json:"created"`
	Img      string    `json:"img" gorm:"type:varchar(255)"`
	Details  string    `json:"details" gorm:"type:varchar(255)"`
}

type Users struct {
	gorm.Model
	Username  string    `json:"username" gorm:"type:varchar(255);unique"`
	Password  string    `json:"password" gorm:"type:varchar(255)"`
	IsStaff   int       `json:"is_staff" gorm:"default:0"`
	LastLogin time.Time `json:"last_login"`
}

type Carts struct {
	gorm.Model
	Quantity    int         `json:"quantity"`
	CommodityId string      `json:"commodity_id"`
	Commodities Commodities `gorm:"foreignkey:CommodityId"`
	UserId      int         `json:"user_id"`
	Users       Users       `json:"-" gorm:"foreignkey:UserId"`
}

type Orders struct {
	gorm.Model
	Price   string `json:"price" gorm:"type:varchar(255)"`
	PayInfo string `json:"payInfo" gorm:"type:varchar(255)"`
	UserId  int64
	Users   Users `json:"-" gorm:"foreignkey:UserId"`
	State   int64 `json:"state"`
}

type Records struct {
	gorm.Model
	CommodityId int64       `json:"CommodityId"`
	Commodities Commodities `grom:"-" gorm:"foreignkey:CommodityId"`
	UserId      int64       `json:"UserId"`
	Users       Users       `json:"-" gorm:"foreignkey:UserId"`
}

type Jwts struct {
	gorm.Model
	Token  string    `json:"token" gorm:"type:varchar(1000)"`
	Expire time.Time `json:"expire"`
}

// 定义模型Users的钩子函数BeforeCrate,为字段Password加密处理
func (u *Users) BeforeSave(db *gorm.DB) error {
	m := md5.New()
	m.Write([]byte(u.Password))
	u.Password = hex.EncodeToString(m.Sum(nil))
	return nil
}

// 定义数据库连接对象
var dsn = fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
	settings.MySQLSetting.User,
	settings.MySQLSetting.Password,
	settings.MySQLSetting.Host,
	settings.MySQLSetting.Name)
var DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
	//禁止创建数据表的外键约束
	DisableForeignKeyConstraintWhenMigrating: true,
})

func Setup() {
	if err != nil {
		fmt.Printf("模型初始化异常：%v", err)
	}
	DB.AutoMigrate(&Types{})
	DB.AutoMigrate(&Commodities{})
	DB.AutoMigrate(&Users{})
	DB.AutoMigrate(&Carts{})
	DB.AutoMigrate(&Jwts{})
	DB.AutoMigrate(&Orders{})
	DB.AutoMigrate(&Records{})

	sqlDB, _ := DB.DB()
	//空闲连接最大数量
	sqlDB.SetMaxIdleConns(10)
	//打开连接最大数量
	sqlDB.SetMaxOpenConns(100)
	//连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)
}
