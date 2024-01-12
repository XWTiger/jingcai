package claim

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jingcai/common"
	alog "jingcai/log"
	"jingcai/mysql"
	"jingcai/order"
	"jingcai/user"
	"jingcai/validatior"
)

var log = alog.Logger

const (
	SCORE = "SCORE"
	RMB   = "RMB"
)

// 兑换
type Claim struct {
	gorm.Model
	//订单id
	OrderId string `validate:"required" message:"订单号必填"`
	//SCORE（积分） / RMB（转账）
	Way string
	//TODO 后期可能区分店铺

	UserId uint

	//是否完成
	Status bool
}

// @Summary 兑换接口
// @Description 兑换接口
// @Tags user  用户
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Claim false "申报对象"
// @Router /api/user/claims [post]
func UserClaim(c *gin.Context) {

	userInfo := user.FetUserInfo(c)
	if userInfo == (user.User{}) {
		common.FailedAuthReturn(c, "用户未登录")
		return
	}
	var claim Claim
	c.BindJSON(&claim)
	if claim.OrderId == "" {
		common.FailedReturn(c, "订单号不能为空")
		return
	}
	claim.Way = SCORE
	claim.UserId = userInfo.ID
	claim.Status = false

	mysql.DB.Save(&claim)

}

type ClaimUpdate struct {
	//兑换id
	ID uint
	//订单id
	OrderId string `validate:"required" message:"订单号必填"`
	//SCORE（积分） / RMB（转账）
	Way string
	//TODO 后期可能区分店铺,只能兑换自己名下的人的票

	UserId uint `validate:"required" message:"用户号必填"`

	//积分
	Score float32 `validate:"required" message:"用户号必填"`
}

type ClaimVO struct {
	Order    *order.OrderVO
	UserInfo *user.User
}

// @Summary 兑换处理接口
// @Description 兑换处理接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Claim false "申报对象"
// @Router /api/super/calims [post]
func UpdateClaim(c *gin.Context) {
	userInfo := user.FetUserInfo(c)
	var update ClaimUpdate
	c.BindJSON(&update)

	err := validatior.Validator(c, update)
	if err != nil {
		log.Error(err)
		return
	}

	if userInfo == (user.User{}) {
		common.FailedAuthReturn(c, "用户未登录")
		return
	}

	var count int64
	tx := mysql.DB.Begin()
	tx.Model(user.User{Model: gorm.Model{
		ID: update.UserId,
	}}).Count(&count)
	if count < 0 {
		common.FailedReturn(c, "用户不存在")
		return
	}

	scoreErr := user.AddScoreInner(update.Score, update.UserId, userInfo.ID, SCORE, tx)
	if scoreErr != nil {
		common.FailedReturn(c, scoreErr.Error())
		return
	} else {
		err := tx.Model(Claim{Model: gorm.Model{ID: update.ID}}).Update("status", true).Error
		if err != nil {
			log.Error(err)
			tx.Rollback()
			common.FailedReturn(c, "兑换失败， 请联系管理员!")
			return
		}
	}
	tx.Commit()
}

// @Summary 兑换查询接口
// @Description 兑换查询接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param page query int false "申报对象"
// @Router /api/super/claims [get]
func ClaimList(c *gin.Context) {
	//TODO 后期只能查自己名下的
	userInfo := user.FetUserInfo(c)
	if userInfo == (user.User{}) {
		common.FailedAuthReturn(c, "用户未登录")
		return
	}
	var claims []Claim
	mysql.DB.Model(Claim{Status: false}).Find(&claims)

	var rep = make([]ClaimVO, 0)
	for _, claim := range claims {

		vo := ClaimVO{}
		ord := order.FindOrderVOById(claim.OrderId, true)
		vo.Order = &ord
		usr := user.FindUserById(claim.UserId)
		vo.UserInfo = &usr
		rep = append(rep, vo)
	}
	common.SuccessReturn(c, rep)

}
