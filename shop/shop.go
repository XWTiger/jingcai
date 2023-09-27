package shop

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/validatior"
	"strconv"
	"strings"
	"time"
)

var log = ilog.Logger
var userCahe = cache2go.Cache("user")

type Shop struct {
	gorm.Model
	//门店地址
	Addr string `validate:"required"`

	UserId uint

	//证件地址
	Certificate string `validate:"required"`

	//经度
	Longitude float32

	//纬度
	Latitude float32

	IdCard string `validate:"required"`

	//门店名称
	Name string `validate:"required"`

	//审核状态
	Status bool

	//用户基本信息
	User user.UserVO `gorm:"-:all"`
}
type Statistics struct {
	//在线人数
	OnlineNum int

	//今日流水
	TodayBill float32

	shop Shop
}

//TODO 店铺注册分享码 注册分享码、网站分享码

// @Summary 店铺注册
// @Description 店铺注册
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Shop false "订单对象"
// @param sharedId query   uint false "开始日期 2023-09-25 00:00:00"
// @Router /api/shop [post]
func ShopRegistry(c *gin.Context) {
	var shopvo Shop
	c.BindJSON(&shopvo)
	var intId int
	id := c.Query("sharedId")
	if strings.Compare(id, "") == 0 {
		intId = 1
	} else {
		ind, err := strconv.Atoi(id)
		if err != nil {
			log.Error(err)
			intId = 1
		} else {
			intId = ind
		}
	}

	err := validatior.Validator(c, shopvo)
	if err != nil {
		log.Error(err)
		return
	}
	tx := mysql.DB.Begin()
	var count int64
	tx.Model(user.User{}).Where(user.User{Phone: shopvo.User.Phone}).Count(&count)
	if count > 0 {
		common.FailedReturn(c, "手机号已经被注册")
		return
	}
	var userPo = user.User{
		Phone:    shopvo.User.Phone,
		Secret:   shopvo.User.Secret,
		Name:     shopvo.Name,
		Salt:     uuid.NewV4().String()[0:16],
		Role:     user.ADMIN,
		Score:    0.00,
		FromUser: uint(intId),
	}

	pwd, err := common.EnPwdCode([]byte(shopvo.User.Secret), []byte(userPo.Salt))
	if err != nil {
		log.Error("加密密码失败", err)
		common.FailedReturn(c, "加密失败，请联系管理员！")
	}
	userPo.Secret = pwd
	tx.Create(&shopvo)
	userPo.From = shopvo.ID
	tx.Create(&userPo)
	tx.Model(Shop{}).Where("id=?", shopvo.ID).Update("user_id", userPo.ID)
	tx.Commit()
	common.SuccessReturn(c, "注册成功,祝老板财源广进!")
}

// @Summary 门店流水
// @Description 门店流水
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param start query   uint false "开始日期 2023-09-25 00:00:00"
// @param end query   uint false "结束日期 2023-09-25 23:59:59"
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @Router /api/super/bills [get]
func ShopBills(c *gin.Context) {
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	start := c.Query("start")
	end := c.Query("end")

	var userInfo = user.FetUserInfo(c)
	var shopInfo Shop
	if err := mysql.DB.Model(Shop{}).Where(&Shop{UserId: userInfo.ID}).First(&shopInfo).Error; err != nil {
		log.Error("该用户没有店铺 user id： ", userInfo.ID)
		common.FailedReturn(c, "您还没有注册店铺")
		return
	}
	if !shopInfo.Status {
		log.Error("该用户店铺没审批通过 user id： ", userInfo.ID)
		common.FailedReturn(c, "您的店铺还没审批通过，请联系管理员!")
		return
	}
	pageN, _ := strconv.Atoi(pageNo)
	pageS, _ := strconv.Atoi(pageSize)
	var bills user.Bill
	mysql.DB.Model(user.Bill{}).Where(user.Bill{ShopId: shopInfo.ID}).Where("created_at BETWEEN ? AND ? ", start, end).Offset((pageN - 1) * pageS).Limit(pageS).Find(&bills)
	common.SuccessReturn(c, bills)
}

// @Summary 管理员基础统计
// @Description 管理员基础统计
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/super/statistics [get]
func StatisticsCount(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	var shopInfo Shop
	if err := mysql.DB.Model(Shop{}).Where(&Shop{UserId: userInfo.ID}).First(&shopInfo).Error; err != nil {
		log.Error("该用户没有店铺 user id： ", userInfo.ID)
		common.FailedReturn(c, "您还没有注册店铺")
		return
	}
	var bills []user.Bill
	year, month, day := time.Now().Date()
	var dateStart = fmt.Sprintf("%d-%d-%d 00:00:00", year, int(month), day)
	var dateEnd = fmt.Sprintf("%d-%d-%d 23:59:59", year, int(month), day)
	mysql.DB.Model(user.Bill{}).Where(user.Bill{ShopId: shopInfo.ID}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Find(&bills)
	var count float32 = 0
	for _, bill := range bills {
		count = count + bill.Num
	}
	var statics = Statistics{
		OnlineNum: userCahe.Count(),
		TodayBill: count,
		shop:      shopInfo,
	}
	common.SuccessReturn(c, statics)

}

// @Summary 查询店内用户信息
// @Description 查询店内用户信息
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @param userId query int   true  "用户id"
// @Router /api/shop/users [get]
func QueryShopUser(c *gin.Context) {
	param := c.Query("userId")
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	var userInfo = user.FetUserInfo(c)
	var shopInfo Shop
	if err := mysql.DB.Model(Shop{}).Where(&Shop{UserId: userInfo.ID}).First(&shopInfo).Error; err != nil {
		log.Error("该用户没有店铺 user id： ", userInfo.ID)
		common.FailedReturn(c, "您还没有注册店铺")
		return
	}

	if param != "" {
		userId, err := strconv.Atoi(param)
		if err != nil {
			log.Error(err)
			common.FailedReturn(c, "用户id 不正确")
			return
		}
		var userInfo user.User
		mysql.DB.Model(user.User{}).Where(&user.User{Model: gorm.Model{
			ID: uint(userId),
		}}).First(&userInfo)
		common.SuccessReturn(c, userInfo)
		return
	} else if pageNo != "" && pageSize != "" {
		pageN, _ := strconv.Atoi(pageNo)
		pageS, _ := strconv.Atoi(pageSize)
		var userInfos []user.User
		mysql.DB.Model(user.User{}).Where(&user.User{From: shopInfo.ID}).Offset((pageN - 1) * pageS).Limit(pageS).Find(&userInfos)
		common.SuccessReturn(c, userInfos)
		return
	}

}
