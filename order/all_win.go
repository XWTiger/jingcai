package order

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"jingcai/validatior"
	"time"
)

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
}

// 合买对象
type AllWinCreate struct {

	//份数
	Number int

	//付款金额
	ShouldPay float32 `max:"0"`

	//支付方式 ALI  WECHAT SCORE（积分）
	PayWay string `validate:"required"`

	//发起合买人订单号
	OrderId string `validate:"required"`

	//发起人
	UserId uint

	//合伙人 id,id
	ParentId uint `validate:"required"`

	//发起成功/失败
	Status bool

	//已经超时
	Timeout bool

	//结束时间
	FinishedTime time.Time

	//购买份数
	BuyNumber int `validate:"required"`
}

func (a AllWin) GetVO() AllWinVO {
	if a.ParentId == 0 {
		var all []AllWin
		var vo = AllWinVO{}
		if err := mysql.DB.Model(AllWin{OrderId: a.OrderId}).Find(&all).Error; err != nil {
			log.Error(err)
			return vo
		}
		order := FindById(a.OrderId)
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
// @Router /order/all_win [get]
func AllWinList(c *gin.Context) {
	var all []AllWin
	mysql.DB.Model(AllWin{
		Timeout:  false,
		ParentId: 0,
	}).Find(&all)
	allVo := make([]AllWinVO, 0)
	for _, win := range all {
		allVo = append(allVo, win.GetVO())
	}
	common.SuccessReturn(c, allVo)
}

// @Summary 合买发起/跟买
// @Description 合买跟买，发起/跟买 自动确认
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Param param body AllWinCreate true "购买对象"
// @Router /order/all_win [post]
func AllWinCreateHandler(c *gin.Context) {
	var body AllWinCreate
	err := c.Bind(&body)
	if err != nil {
		common.FailedReturn(c, "参数解释失败！")
		return
	}
	validatior.Validator(c, body)
	var user = user.FetUserInfo(c)
	tx := mysql.DB.Begin()
	if len(body.OrderId) > 0 {
		//合买
		var order Order

		if err := tx.Model(Order{UUID: body.OrderId}).First(&order).Error; err != nil {
			log.Error("查询发起人订单失败", body.OrderId)
			common.FailedReturn(c, "查询发起人订单失败")
			return
		}
		if order.UserID == body.UserId {
			//发起合买
			var initAllWin = AllWin{
				Timeout:      false,
				FinishedTime: common.GetMatchFinishedTime(order.Matches[0].TimeDate),
				ParentId:     0,
				UserId:       order.UserID,
				OrderId:      order.UUID,
				Number:       body.Number,
				BuyNumber:    body.BuyNumber,
				ShouldPay:    float32(order.ShouldPay/float32(body.Number)) * float32(body.BuyNumber),
				Bonus:        0,
			}
			tx.Save(&initAllWin)
			if err := tx.Model(&Order{UUID: order.UUID}).Update("all_win_id", initAllWin.ID).Error; err != nil {
				log.Error(err)
				log.Error("合买，更新订单的合买id 失败")
				common.FailedReturn(c, "合买，更新订单的合买id 失败")
				tx.Rollback()
				return
			}
		} else {
			//跟买
			var initAll AllWin
			if err := tx.Model(AllWin{Model: gorm.Model{ID: order.AllWinId}}).First(&initAll).Error; err != nil {
				log.Error("查询发起人合买订单失败", body.OrderId)
				common.FailedReturn(c, "查询发起人合买订单失败")
				return
			}
			var userOrder = Order{
				CreatedAt:        time.Now(),
				UUID:             body.OrderId,
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
				UserID:           user.ID,
				ShouldPay:        float32(order.ShouldPay/float32(initAll.Number)) * float32(body.BuyNumber),
				Bonus:            order.Bonus,
				PayWay:           body.PayWay,
				AllMatchFinished: order.AllMatchFinished,
			}

			tx.Save(&userOrder)
			var allWin = AllWin{
				Timeout:      false,
				FinishedTime: common.GetMatchFinishedTime(order.Matches[0].TimeDate),
				ParentId:     initAll.ID,
				UserId:       order.UserID,
				OrderId:      userOrder.UUID,
				Number:       initAll.Number,
				BuyNumber:    body.BuyNumber,
				ShouldPay:    userOrder.ShouldPay,
				Bonus:        0,
			}
			tx.Save(&allWin)
			tx.Model(Order{UUID: userOrder.UUID}).Update("all_win_id", allWin.ID)
		}
	} else {
		//发起合买
		log.Info("=========发起人订单id 为空===========")
		common.FailedReturn(c, "发起人订单id 为空")
		tx.Rollback()
		return
	}
	tx.Commit()

}

func GetAllWinByParentId(parentId uint) []AllWin {
	var all []AllWin
	mysql.DB.Model(AllWin{
		Timeout:  false,
		ParentId: parentId,
	}).Find(&all)
	var init AllWin
	mysql.DB.Model(AllWin{Model: gorm.Model{
		ID: parentId,
	}}).First(&init)
	all = append(all, init)
	return all
}
