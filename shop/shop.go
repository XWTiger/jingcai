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
	user.UserVO `gorm:"-:all"`
}
type Statistics struct {
	//在线人数
	OnlineNum int

	//今日流水
	TodayBill float32

	shop Shop
}

//TODO 店铺注册分享码 注册分享码、网站分享码

// @Summary 订单创建接口
// @Description 订单创建接口， matchs 比赛按时间从先到后排序
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Shop false "订单对象"
// @param sharedId query   uint false "开始日期 2023-09-25 00:00:00"
// @Router /api/super/shop [post]
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

	validatior.Validator(c, shopvo)
	tx := mysql.DB.Begin()
	var userPo = user.User{
		Phone:  shopvo.Phone,
		Secret: shopvo.Secret,
		Name:   shopvo.Name,
		Salt:   uuid.NewV4().String()[0:16],
		Role:   user.ADMIN,
		Score:  0.00,
		From:   uint(intId),
	}

	pwd, err := common.EnPwdCode([]byte(shopvo.Secret), []byte(userPo.Salt))
	if err != nil {

		log.Error("加密密码失败", err)
		common.FailedReturn(c, "加密失败，请联系管理员！")
	}
	userPo.Secret = pwd
	tx.Create(&userPo)
	tx.Create(&shopvo)
	tx.Commit()
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
