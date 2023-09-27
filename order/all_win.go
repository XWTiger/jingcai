package order

import (
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/util"
	"jingcai/validatior"
	"strconv"
	"strings"
	"time"
)

const FOLLOW = "FOLLOW"
const MASTER = "MASTER"

// 合买
type AllWin struct {
	gorm.Model
	//份数
	Number int

	//关联Order id
	OrderId string

	//发起人
	UserId uint

	//合伙人 id,id
	ParentId uint

	ParentOrderId string

	//发起成功/失败
	Status bool

	//已经超时
	Timeout bool

	//结束时间
	FinishedTime time.Time

	//购买份数
	BuyNumber int

	//应付
	ShouldPay float32

	//奖金
	Bonus float32

	//SHARE(公开) AFTER_END(截至后公开)  JOIN（购买后可见）
	ShowType string

	//备注
	Comment string

	//保底份数
	LeastTimes int
}
type AllWinVO struct {
	//份数
	Number int

	//关联发起人Order
	Order Order
	//合伙人 0 为发起人
	Partners []AllWinUser

	//发起成功/失败
	Status bool

	//已经超时
	Timeout bool
	//结束时间
	FinishedTime time.Time

	//SHARE(公开) AFTER_END(截至后公开)  JOIN（购买后可见）
	ShowType string

	// 合买红单率
	WinRate float32

	// 回报率
	ReturnRate float32

	// 带红人数
	FollowWinCount int

	//备注
	Comment string

	//保底份数
	LeastTimes int
}
type AllWinUser struct {
	Phone string
	//昵称
	Name string

	Role string //"enum: Admin,User"
	//微信号
	Wechat string
	//支付宝号
	Ali string

	//余额
	Score float32

	//头像地址
	HeaderImageUrl string

	//份数
	BuyNumber int

	//SHARE(公开) AFTER_END(截至后公开)  JOIN（购买后可见）
	ShowType string
}

// 合买对象
type AllWinCreate struct {

	//份数
	Number int

	//付款金额(总)
	ShouldPay float32 `max:"0"`

	//支付方式 ALI  WECHAT SCORE（积分）
	PayWay string `validate:"required"`

	//发起合买人订单号
	OrderId string `validate:"required"`

	//发起人
	UserId uint

	//发起人是0
	ParentId uint

	//前端不用填， 发起成功/失败
	Status bool

	//前端你不用填， 已经超时
	Timeout bool

	//结束时间 2006-01-02T15:04:05+07:00 2023-08-12T21:20+08:00
	FinishedTime time.Time

	//购买份数
	BuyNumber int `validate:"required"`

	//发起合买时传入该参数，SHARE(公开) AFTER_END(截至后公开)  JOIN（购买后可见）
	ShowType string
	//FOLLOW(跟买)  MASTER(发起)
	BuyType string

	//备注
	Comment string

	//出票保底份数
	LeastTimes int
}

func (a AllWin) GetVO() AllWinVO {
	if a.ParentId == 0 {

		var all []AllWin
		var vo = AllWinVO{}
		if err := mysql.DB.Model(AllWin{OrderId: a.OrderId}).Where(AllWin{OrderId: a.OrderId}).Find(&all).Error; err != nil {
			log.Error(err)
			return vo
		}
		//SHARE(公开) AFTER_END(截至后公开)  JOIN（购买后可见）
		var order Order
		if strings.Compare(a.ShowType, "SHARE") == 0 {
			order = FindById(a.OrderId, true)
		}
		vo.Order = order
		var allW = make([]AllWin, 0)
		var partner = make([]AllWinUser, 0)
		allW = append(allW, a)
		userInfo := GetAllWinUser(user.FindUserById(a.UserId))

		userInfo.BuyNumber = a.BuyNumber
		partner = append(partner, userInfo)
		for _, win := range all {
			userInfos := GetAllWinUser(user.FindUserById(win.UserId))
			userInfos.BuyNumber = win.BuyNumber
			partner = append(partner, userInfos)
		}
		vo.Partners = partner
		vo.Timeout = a.Timeout
		vo.FinishedTime = a.FinishedTime
		vo.Status = a.Status
		vo.LeastTimes = a.LeastTimes
		vo.ShowType = a.ShowType
		vo.Number = a.Number
		//计算合买带红人数
		//1.查到这人所有中奖单
		var allOfThePerson []AllWin
		if err := mysql.DB.Model(AllWin{}).Where("user_id=?", a.UserId).Where("bonus > 0").Find(&allOfThePerson).Error; err != nil {
			log.Warn("查询这个用户合买所有发布中奖列表失败")
			return vo
		}
		var orderIds = make([]string, 0)
		var totalWinMoney float32 = 0.0
		for _, person := range allOfThePerson {
			orderIds = append(orderIds, person.OrderId)
			totalWinMoney += person.Bonus
		}
		var count int64
		mysql.DB.Model(AllWin{}).Where("order_id in (?)", orderIds).Count(&count)
		vo.FollowWinCount = int(count)
		//计算合买红单率
		var totalAllOfPerson []AllWin
		if err := mysql.DB.Model(AllWin{}).Where("user_id=?", a.UserId).Find(&totalAllOfPerson).Error; err != nil {
			log.Warn("查询这个用户所有发布合买失败")
			return vo
		}
		vo.WinRate = float32(len(allOfThePerson)) / float32(len(totalAllOfPerson))
		//计算回报率
		var wasteMoney float32 = 0.0
		for _, person := range totalAllOfPerson {
			wasteMoney += person.ShouldPay
		}
		vo.ReturnRate = totalWinMoney / wasteMoney
		vo.Comment = a.Comment

		return vo
	} else {
		return AllWinVO{}
	}
}

func GetAllWinUser(u user.User) AllWinUser {
	return AllWinUser{
		Phone:          u.Phone,
		Name:           u.Name,
		Role:           u.Role,
		Wechat:         u.Wechat,
		Ali:            u.Ali,
		Score:          u.Score,
		HeaderImageUrl: u.HeaderImageUrl,
	}
}

// @Summary 合买列表
// @Description 合买列表
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @param lotteryType  query string false "足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）"
// @Router /api/order/all_win [get]
func AllWinList(c *gin.Context) {
	param := c.Query("lotteryType")
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")

	var all []AllWin
	var list []Order
	var pageN, pageS int
	allVo := make([]AllWinVO, 0)
	if len(param) > 0 {
		if pageNo == "" || pageSize == "" {
			mysql.DB.Model(Order{}).Where("lottery_type=? and all_win_id > 0", param).Find(&list)
		} else {
			pageN, _ = strconv.Atoi(pageNo)
			pageS, _ = strconv.Atoi(pageSize)
			mysql.DB.Model(Order{}).Where("lottery_type=? and all_win_id > 0", param).Offset((pageN - 1) * pageS).Limit(pageS).Find(&list)
		}

		if len(list) <= 0 {
			common.SuccessReturn(c, allVo)
			return
		}
		var allWinIds = make([]uint, 0)
		for _, order := range list {
			allWinIds = append(allWinIds, order.AllWinId)
		}
		mysql.DB.Model(AllWin{}).Where(&AllWin{
			Timeout:  false,
			ParentId: 0,
		}).Where("id in (?)", allWinIds).Find(&all)
	} else {
		mysql.DB.Model(AllWin{}).Where(&AllWin{
			Timeout:  false,
			ParentId: 0,
		}).Find(&all)
	}
	var count int64
	mysql.DB.Model(Order{}).Where("lottery_type=? and all_win_id > 0", param).Count(&count)
	for _, win := range all {
		allVo = append(allVo, win.GetVO())
	}
	common.SuccessReturn(c, common.PageCL{
		PageNo:   pageN,
		PageSize: pageS,
		Total:    int(count),
		Content:  allVo,
	})
}

// @Summary 合买发起/跟买
// @Description 合买跟买，发起/跟买 自动确认
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Param param body AllWinCreate true "购买对象"
// @Router /api/order/all_win [post]
func AllWinCreateHandler(c *gin.Context) {
	var body AllWinCreate
	err := c.Bind(&body)
	if err != nil {
		log.Error(err)
		common.FailedReturn(c, "参数解释失败！")
		return
	}
	validatior.Validator(c, body)
	var userInfo = user.FetUserInfo(c)
	tx := mysql.DB.Begin()
	if len(body.OrderId) > 0 {
		//合买
		var order Order

		if err := tx.Model(Order{UUID: body.OrderId}).Where(&Order{UUID: body.OrderId}).First(&order).Error; err != nil {
			log.Error("查询发起人订单失败", body.OrderId)
			common.FailedReturn(c, "查询发起人订单失败")
			return
		}
		if strings.Compare(order.LotteryType, FOOTBALL) == 0 || strings.Compare(order.LotteryType, BASKETBALL) == 0 {
			var matchs []Match
			if err := tx.Model(&Match{}).Where(&Match{OrderId: order.UUID}).Find(&matchs).Error; err != nil {
				log.Error(err)
				common.FailedReturn(c, "查询发起人订单失败")
				return
			}
			order.Matches = matchs
		}

		if strings.Compare(body.BuyType, MASTER) == 0 {
			//发起合买
			var initAllWin = AllWin{
				Timeout:      false,
				FinishedTime: getFinishedTime(order),
				ParentId:     0,
				UserId:       order.UserID,
				OrderId:      order.UUID,
				Number:       body.Number,
				BuyNumber:    body.BuyNumber,
				ShouldPay:    float32(order.ShouldPay/float32(body.Number)) * float32(body.BuyNumber),
				Bonus:        0,
				ShowType:     body.ShowType,
				Comment:      body.Comment,
				LeastTimes:   body.LeastTimes,
			}
			if initAllWin.BuyNumber > initAllWin.LeastTimes {
				common.FailedReturn(c, "保底份数不能小于认购份数")
				return
			}
			var shouldPay = float32(order.ShouldPay/float32(body.Number)) * float32(initAllWin.LeastTimes)
			leastErr := user.CheckScoreOrDoBill(initAllWin.UserId, shouldPay, false, tx)
			if leastErr != nil {
				common.FailedReturn(c, "积分不够付款保底份数")
				tx.Rollback()
				return
			}
			payErr := user.CheckScoreOrDoBill(initAllWin.UserId, initAllWin.ShouldPay, true, tx)
			if payErr != nil {
				log.Error("all win id: ", initAllWin.ID, "扣款失败")
				tx.Rollback()
				common.FailedReturn(c, payErr.Error())
				return
			}
			tx.Save(&initAllWin)
			if err := tx.Model(&Order{UUID: order.UUID}).Where(&Order{UUID: order.UUID}).Update("all_win_id", initAllWin.ID).Error; err != nil {
				log.Error(err)
				log.Error("合买，更新订单的合买id 失败")
				common.FailedReturn(c, "合买，更新订单的合买id 失败")
				tx.Rollback()
				return
			}
			jobTime := initAllWin.FinishedTime
			AllWinCheck(jobTime)
		} else {
			//跟买
			if body.BuyNumber <= 0 {
				log.Error("合买份数不能小于0", body.OrderId)
				common.FailedReturn(c, "合买份数不能小于0")
				return
			}
			var initAll AllWin
			if err := tx.Model(AllWin{}).Where(&AllWin{Model: gorm.Model{ID: order.AllWinId}}).First(&initAll).Error; err != nil {
				log.Error("查询发起人合买订单失败", body.OrderId)
				common.FailedReturn(c, "查询发起人合买订单失败")
				return
			}
			//校验是否还能合买，如果自己认购 + 别人认购 < 总数可以购买
			var allOrder = make([]AllWin, 0)
			tx.Model(AllWin{}).Where(&AllWin{ParentId: initAll.ParentId}).Find(&allOrder)
			var count = 0
			for _, win := range allOrder {
				count += win.BuyNumber
			}
			times := count + initAll.BuyNumber + body.BuyNumber
			if times > initAll.Number {
				log.Error("合买份数总和不能大于", initAll.Number)
				common.FailedReturn(c, fmt.Sprintf("合买份数不能大于%d", initAll.Number-count-initAll.BuyNumber))
				return
			}
			var userOrder = Order{
				CreatedAt:        time.Now(),
				UUID:             uuid.NewV4().String(),
				Times:            order.Times,
				Way:              order.Way,
				LotteryType:      order.LotteryType,
				LogicWinMin:      order.LogicWinMin,
				LogicWinMaX:      order.LogicWinMaX,
				LotteryUuid:      order.LotteryUuid,
				Content:          order.Content,
				SaveType:         order.SaveType,
				Share:            false,
				AllWinId:         0,
				UserID:           userInfo.ID,
				ShouldPay:        float32(order.ShouldPay/float32(initAll.Number)) * float32(body.BuyNumber),
				Bonus:            order.Bonus,
				PayWay:           body.PayWay,
				AllMatchFinished: order.AllMatchFinished,
			}
			payErr := user.CheckScoreOrDoBill(userInfo.ID, userOrder.ShouldPay, true, tx)
			if payErr != nil {
				log.Error("user id: ", userInfo.ID, "扣款失败")
				tx.Rollback()
				common.FailedReturn(c, payErr.Error())
				return
			}
			tx.Save(&userOrder)
			var allWin = AllWin{
				Timeout:       false,
				FinishedTime:  common.GetMatchFinishedTime(order.Matches[0].TimeDate),
				ParentId:      initAll.ID,
				UserId:        order.UserID,
				OrderId:       userOrder.UUID,
				Number:        initAll.Number,
				BuyNumber:     body.BuyNumber,
				ShouldPay:     userOrder.ShouldPay,
				Bonus:         0,
				ShowType:      initAll.ShowType,
				ParentOrderId: order.UUID,
			}
			if count+initAll.LeastTimes >= initAll.Number {
				allWin.Status = true
				for i := 0; i < len(allOrder); i++ {
					allOrder[i].Status = true
				}
				initAll.Status = true
				allWin.Status = true
				tx.Save(initAll)
				tx.Save(allOrder)
			}
			tx.Save(&allWin)
			if initAll.Status {
				tx.Model(Order{UUID: userOrder.UUID}).Where(&Order{UUID: userOrder.UUID}).Update("pay_status", true).Update("all_win_id", allWin.ID)
			} else {
				tx.Model(Order{UUID: userOrder.UUID}).Where(&Order{UUID: userOrder.UUID}).Update("all_win_id", allWin.ID)
			}
		}
	} else {
		//发起合买
		log.Info("=========发起人订单id 为空===========")
		common.FailedReturn(c, "发起人订单id 为空")
		tx.Rollback()
		return
	}
	tx.Commit()
	common.SuccessReturn(c, body)
	return

}

func GetAllWinByParentId(parentId uint) []AllWin {
	var all []AllWin
	mysql.DB.Model(AllWin{
		Timeout:  false,
		ParentId: parentId,
	}).Find(&all)
	var init AllWin
	mysql.DB.Debug().Model(AllWin{Model: gorm.Model{
		ID: parentId,
	}}).First(&init)
	all = append(all, init)
	return all
}

func getFinishedTime(order Order) time.Time {
	switch order.LotteryType {

	case FOOTBALL:
	case BASKETBALL:
		return common.GetMatchFinishedTime(order.Matches[0].TimeDate)
	case P3:
	case P5:
		return util.GetPLWFinishedTime()
	case SUPER_LOTTO:

		now := time.Now()
		time := util.GetPLWFinishedTime()
		if now.Weekday() == 1 || now.Weekday() == 3 || now.Weekday() == 6 {
			return util.GetPLWFinishedTime()
		} else {
			return time.AddDate(0, 0, 1)
		}

	case SEVEN_STAR:
		now := time.Now()
		time := util.GetPLWFinishedTime()
		if now.Weekday() == 2 || now.Weekday() == 5 || now.Weekday() == 0 {
			return util.GetPLWFinishedTime()
		} else {
			return time.AddDate(0, 0, 1)
		}

	default:
		log.Error("获取合买结束时间失败！返回当前时间")
		return time.Now()
	}
	return time.Now()
}

//  合买是否保底校验校验, 如果其他人购买数 + 自己认购 不等于 总份数 并且已经超时

func AllWinCheck(when time.Time) {

	job := Job{
		Time: when,
		CallBack: func(param interface{}) {
			var allWinOrders = make([]AllWin, 0)
			mysql.DB.Model(AllWin{}).Where(AllWin{Timeout: false, Status: false, ParentId: 0}).Find(&allWinOrders)
			if len(allWinOrders) > 0 {
				for i := 0; i < len(allWinOrders); i++ {
					var allWin = allWinOrders[i]

					var partners = make([]AllWin, 0)
					mysql.DB.Model(AllWin{}).Where(&AllWin{ParentId: allWin.ID}).Find(partners)
					if len(partners) > 0 {
						var count = 0
						for _, partner := range partners {
							count += partner.BuyNumber
						}
						if allWin.FinishedTime.Second()-time.Now().Second() < 0 {
							allWin.Timeout = true
						}
						var number = allWin.Number - (count + allWin.BuyNumber)
						if allWin.LeastTimes > number && number > 0 && allWin.FinishedTime.Second()-time.Now().Second() < 0 {
							//到期，需要保底 退换多扣的
							shouldReturn := float32(allWin.LeastTimes-number) * float32(allWin.ShouldPay/float32(allWin.Number))
							returnErr := user.ReturnScore(allWin.UserId, shouldReturn)
							if returnErr != nil {
								log.Error("user id: ", allWin.UserId, "退还失败, 金额：", shouldReturn, "订单号：", allWin.ParentOrderId)
								continue
							}
							allWin.Status = true
						}
						if number == 0 {
							//发起成功
							allWin.Status = true
						}
						if allWin.Status == true {
							for _, partner := range partners {
								partner.Status = true
								mysql.DB.Model(AllWin{}).Save(partner)
							}
						}
					} else {
						if allWin.FinishedTime.Second()-time.Now().Second() < 0 {
							allWin.Timeout = true
						}
					}
					if err := mysql.DB.Model(AllWin{}).Save(allWin).Error; err != nil {
						log.Error(err, "更新发起者状态失败")
						continue
					}

				}

			}
		},
	}
	AddJob(job)
}
