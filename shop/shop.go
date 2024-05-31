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
	"jingcai/score"
	"jingcai/user"
	"jingcai/validatior"
	"strconv"
	"strings"
	"time"
)

var log = ilog.Logger
var userCahe = cache2go.Cache("user")

const (
	ADD      = "ADD"      //入账
	SUBTRACT = "SUBTRACT" //出账
)

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

// 统计对象
type Statistics struct {

	//总用户数
	TotalUserNum int64

	//在线人数
	OnlineNum int

	//今日流水
	TodayBill float32

	//积分
	Score float32

	//赠送积分
	FreeScore float32

	//用户账户总积分
	UserTotalScore float32

	//用户账户免费积分
	UserTotalFreeScore float32

	shop Shop
}

//TODO 店铺注册分享码 注册分享码、网站分享码

// @Summary 店铺注册
// @Description 店铺注册
// @Tags shop 门店
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
// @Tags owner 店主
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
	var bills []user.Bill
	var count int64
	if start != "" && end != "" {
		mysql.DB.Model(user.Bill{}).Where(user.Bill{ShopId: shopInfo.ID}).Where("created_at BETWEEN ? AND ? ", start, end).Count(&count).Offset((pageN - 1) * pageS).Limit(pageS).Order("created_at desc").Find(&bills)
	} else {
		mysql.DB.Debug().Model(user.Bill{}).Where(user.Bill{ShopId: shopInfo.ID}).Count(&count).Offset((pageN - 1) * pageS).Limit(pageS).Order("created_at desc").Find(&bills)
	}

	common.SuccessReturn(c, common.PageCL{
		PageNo:   pageN,
		PageSize: pageS,
		Total:    int(count),
		Content:  bills,
	})
}

// @Summary 管理员基础统计
// @Description 管理员基础统计
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} Statistics
// @failure 500 {object} common.BaseResponse
// @param start query   uint true "开始日期 2023-09-25 00:00:00"
// @param end query   uint true "结束日期 2023-09-25 23:59:59"
// @param type query  string false "BILL_COMMENT_CASHED(兑奖),BILL_COMMENT_ADD(充值),BILL_COMMENT_CLEAR(清账),BILL_COMMENT_ACTIVITY(活动赠送),BILL_COMMENT_BUY(购彩)"
// @Param option query string false "option ADD(入账),SUBTRACT(出账) option 和type 不能同时传 二选一"
// @Router /api/super/statistics [get]
func StatisticsCount(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	var shopInfo Shop
	typ := c.Query("type")
	opt := c.Query("option")
	start := c.Query("start")
	end := c.Query("end")
	if err := mysql.DB.Model(Shop{}).Where(&Shop{UserId: userInfo.ID}).First(&shopInfo).Error; err != nil {
		log.Error("该用户没有店铺 user id： ", userInfo.ID)
		common.FailedReturn(c, "您还没有注册店铺")
		return
	}
	var params = user.Bill{ShopId: shopInfo.ID, Comment: user.BILL_COMMENT_BUY}
	var bills []user.Bill
	year, month, day := time.Now().Date()
	var dateStart = fmt.Sprintf("%d-%d-%d 00:00:00", year, int(month), day)
	var dateEnd = fmt.Sprintf("%d-%d-%d 23:59:59", year, int(month), day)
	if opt != "" {
		switch opt {
		case ADD:
			params.Option = ADD
			break
		case SUBTRACT:
			params.Option = SUBTRACT
			break

		}
	} else {
		if typ != "" {
			switch typ {
			case user.BILL_COMMENT_BUY:
				params.Comment = user.BILL_COMMENT_BUY
				break
			case user.BILL_COMMENT_CLEAR:
				params.Comment = user.BILL_COMMENT_CLEAR
				break
			case user.BILL_COMMENT_ADD:
				params.Comment = user.BILL_COMMENT_ADD
				break
			case user.BILL_COMMENT_CASHED:
				params.Comment = user.BILL_COMMENT_CASHED
				break
			case user.BILL_COMMENT_ACTIVITY:
				params.Comment = user.BILL_COMMENT_ACTIVITY
				break
			}
		}
	}
	if start != "" && end != "" {
		dateStart = start
		dateEnd = end
	}
	mysql.DB.Model(user.Bill{}).Where(&params).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Find(&bills)
	var count float32 = 0

	var freeScore float32
	var scoreNum float32

	for _, bill := range bills {
		count = count + bill.Num
		if bill.Type == user.FREE_SCORE {
			freeScore += bill.Num
		}
		if bill.Type == user.SCORE {
			scoreNum += bill.Num
		}
	}
	var totalUser int64
	var users []user.User
	mysql.DB.Model(&user.User{}).Count(&totalUser).Find(&users)

	var userTotalScore float32
	var userTotalFreeScore float32
	for _, u := range users {
		userTotalScore += u.Score
		free, err := score.QueryByUserId(u.ID)
		if err == nil {
			userTotalFreeScore += free.Score
		}
	}

	var statics = Statistics{
		TotalUserNum:       totalUser,
		UserTotalFreeScore: userTotalFreeScore,
		UserTotalScore:     userTotalScore,
		OnlineNum:          userCahe.Count(),
		TodayBill:          count,
		Score:              scoreNum,
		FreeScore:          freeScore,
		shop:               shopInfo,
	}
	common.SuccessReturn(c, statics)

}

// 统计对象 入账： 充值、兑奖、活动赠送； 出账：购彩，清账
// "BILL_COMMENT_CASHED":   "兑奖",
//
//	"BILL_COMMENT_ADD":      "充值",
//	"BILL_COMMENT_CLEAR":    "清账",
//	"BILL_COMMENT_ACTIVITY": "活动赠送",
//	"BILL_COMMENT_BUY":      "购彩",
type BillStatistics struct {
	FreeScoreCount float32
	ScoreCount     float32
	Bills          []user.BillVO
}

// @Summary 门店流水丰富查询
// @Description 门店流水丰富查询 统计对象 入账： 充值、兑奖、活动赠送； 出账：购彩，清账
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param start query   uint true "开始日期 2023-09-25 00:00:00"
// @param end query   uint true "结束日期 2023-09-25 23:59:59"
// @param type query  string false "BILL_COMMENT_CASHED(兑奖),BILL_COMMENT_ADD(充值),BILL_COMMENT_CLEAR(清账),BILL_COMMENT_ACTIVITY(活动赠送),BILL_COMMENT_BUY(购彩)"
// @Param option query string false "option ADD(入账),SUBTRACT(出账) option 和type 不能同时传 二选一"
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @Router /api/shop/bills/all [get]
func ShopOwnerComprehensiveness(c *gin.Context) {
	opt := c.Query("option")
	start := c.Query("start")
	end := c.Query("end")
	typ := c.Query("type")
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	pageN, _ := strconv.Atoi(pageNo)
	pageS, _ := strconv.Atoi(pageSize)
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
	var bills []user.Bill
	var params = user.Bill{ShopId: shopInfo.ID}
	if opt != "" {
		switch opt {
		case ADD:
			params.Option = ADD
			break
		case SUBTRACT:
			params.Option = SUBTRACT
			break

		}
	} else {
		if typ != "" {
			switch typ {
			case user.BILL_COMMENT_BUY:
				params.Comment = user.BILL_COMMENT_BUY
				break
			case user.BILL_COMMENT_CLEAR:
				params.Comment = user.BILL_COMMENT_CLEAR
				break
			case user.BILL_COMMENT_ADD:
				params.Comment = user.BILL_COMMENT_ADD
				break
			case user.BILL_COMMENT_CASHED:
				params.Comment = user.BILL_COMMENT_CASHED
				break
			case user.BILL_COMMENT_ACTIVITY:
				params.Comment = user.BILL_COMMENT_ACTIVITY
				break
			}
		}
	}
	var count int64
	mysql.DB.Model(user.Bill{}).Where(&params).Where("created_at BETWEEN ? AND ? ", start, end).Count(&count).Offset((pageN - 1) * pageS).Limit(pageS).Find(&bills)
	/*var freeScore float32
	var score float32*/
	var vos []user.BillVO
	for _, bill := range bills {
		/*if bill.Type == user.FREE_SCORE {
			freeScore += bill.Num
		}
		if bill.Type == user.SCORE {
			score += bill.Num
		}*/
		vos = append(vos, *bill.GetVO())
	}
	common.SuccessReturn(c, common.PageCL{
		PageNo:   pageN,
		PageSize: pageS,
		Total:    int(count),
		Content:  vos,
	})
}

// @Summary 查询店内用户信息
// @Description 查询店内用户信息
// @Tags shop 门店
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @param userId query int   false  "用户id"
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
		dto := userInfo.GetDTO()
		free, _ := score.QueryByUserId(dto.ID)
		dto.FreeScore = free.Score
		common.SuccessReturn(c, dto)
		return
	} else if pageNo != "" && pageSize != "" {
		pageN, _ := strconv.Atoi(pageNo)
		pageS, _ := strconv.Atoi(pageSize)
		var userInfos []user.User
		var count int64
		mysql.DB.Model(user.User{}).Where(&user.User{From: shopInfo.ID}).Count(&count).Offset((pageN - 1) * pageS).Limit(pageS).Find(&userInfos)
		common.SuccessReturn(c, common.PageCL{
			pageN,
			pageS,
			int(count),
			userInfos,
		})
		return
	}

}
