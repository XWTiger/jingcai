package order

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jingcai/cache"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"strconv"
	"strings"
)

type OrderImage struct {
	gorm.Model
	//访问地址
	Url string
	//是否删除
	Delete bool

	//order uuid
	ParentId string
}

func getImageByOrderId(oid string) []OrderImage {

	var images = make([]OrderImage, 0)
	if err := mysql.DB.Model(OrderImage{}).Where(&OrderImage{ParentId: oid}).Find(&images).Error; err != nil {
		log.Error(err)
		return images
	}
	return images
}

type UploadBet struct {
	//订单id
	OrderId string `validate:"required"`

	//图片地址
	Url []string `validate:"required"`

	//查询id
	LotteryUuid string `validate:"required" gorm:"-:all"`

	//赔率是否和购买赔率不一致， 不一致就
	OddChange bool

	//如果OddChange 为true 才传该对象
	MatchOdds []MatchOdd
}

type MatchOdd struct {
	MatchId string `validate:"required"`
	//足球类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）
	Type string `validate:"required"`
	//让球 胜平负才有，篮球就是让分
	GoalLine string

	//赔率
	Odds float32
}

// @Summary 票提交接口
// @Description 票提交接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param bets  body UploadBet true "管理员提交票对象"
// @Router /api/super/bets [post]
func UploadBets(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	if strings.Compare(userInfo.Role, user.ADMIN) != 0 {
		common.FailedReturn(c, "该接口只提供给管理员")
		return
	}
	var betImgObj UploadBet
	err := c.BindJSON(&betImgObj)
	if err != nil {
		common.FailedReturn(c, "参数错误")
		return
	}
	tx := mysql.DB.Begin()
	var order Order
	if ordErr := tx.Model(&Order{}).Where(&Order{UUID: betImgObj.OrderId}).First(&order).Error; ordErr != nil {
		common.FailedReturn(c, "订单查询失败")
		return
	}
	//校验
	switch order.LotteryType {
	case FOOTBALL:
		_, err := cache.GetOnTimeFootballMatch(betImgObj.LotteryUuid)
		if err == nil {
			common.FailedReturn(c, "非法参数")
			return
		}
		break
	case BASKETBALL:
		_, err := cache.GetOnTimeBasketBallMatch(order.LotteryUuid)
		if err == nil {
			common.FailedReturn(c, "非法参数")
			return
		}
		break

	}
	//调整赔率
	if betImgObj.OddChange && len(betImgObj.MatchOdds) > 0 {
		var matchs = betImgObj.MatchOdds

		for i := 0; i < len(matchs); i++ {
			var matchPo Match
			if err := tx.Model(Match{}).Where(&Match{
				MatchId: matchs[i].MatchId,
			}).Find(&matchPo).Error; err != nil {
				log.Error(err)
				log.Error("查不到订单比赛记录")
				common.FailedReturn(c, "查不到比赛记录！！")
				tx.Rollback()
				return
			}
			var lottery LotteryDetail
			if lerr := tx.Model(LotteryDetail{}).Where(&LotteryDetail{ParentId: matchPo.ID}).Find(&lottery).Error; lerr != nil {
				log.Error("更新赔率 没有查询到票的信息")
				common.FailedReturn(c, "更新赔率 没有查询到票的信息！！")
				tx.Rollback()
				return
			}
			tx.Save(lottery)
		}
	}
	//保存图片
	if len(betImgObj.Url) <= 0 {
		common.FailedReturn(c, "图片信息为空")
		return
	}
	var images []OrderImage
	for _, s := range betImgObj.Url {
		images = append(images, OrderImage{
			Url:      s,
			ParentId: order.UUID,
		})
	}
	if imageErr := tx.Create(&images).Error; imageErr != nil {
		common.FailedReturn(c, "图标保存失败")
		tx.Rollback()
		return
	}
	tx.Commit()
}

// @Summary 管理员订单查询接口
// @Description 订单查询接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param saveType  query string false "保存类型 TEMP（临时保存） TOMASTER（提交到店）  ALLWIN（合买）"
// @param lotteryType  query string false "足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）"
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @Router /api/super/order [get]
func AdminOrderList(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	saveType := c.Query("saveType")
	lotteryType := c.Query("lotteryType")
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	// TODO
	if strings.Compare(userInfo.Role, user.ADMIN) != 0 {
		common.FailedReturn(c, "该接口只提供给管理员")
		return
	}
	var param = Order{
		SaveType:    saveType,
		LotteryType: lotteryType,
	}
	var list = make([]Order, 0)
	var resultList = make([]*OrderVO, 0)
	query := mysql.DB.Model(param).Where("all_win_id > 0 and pay_status = true and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')").Joins("INNER JOIN all_wins on orders.all_win_id = all_wins.id and all_wins.`parent_id`=0  and all_wins.`status` = true")
	query2 := mysql.DB.Model(param).Select("orders.* ").Where("orders.deleted_at is null and pay_status = true and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')")
	var count int64
	mysql.DB.Raw("? union ? ", query, query2).Count(&count)
	mysql.DB.Raw("? union ? ", query, query2).Order("orders.create_at desc").Offset((page - 1) * pageSize).Limit(pageSize).Scan(&list)

	for index, order := range list {
		var mathParam = Match{
			OrderId: list[index].UUID,
		}
		var matchList = make([]Match, 0)
		mysql.DB.Model(&mathParam).Find(&matchList)
		list[index].Matches = matchList
		for _, match := range matchList {
			var detailParam = LotteryDetail{
				ParentId: match.ID,
			}
			var detailList = make([]LotteryDetail, 0)
			mysql.DB.Model(&detailParam).Find(&detailList)
			match.Combines = detailList
		}
		resultList = append(resultList, &OrderVO{
			Order:  &list[index],
			Images: getImageByOrderId(order.UUID),
		})
	}

	common.SuccessReturn(c, &common.PageCL{
		PageNo:   page,
		PageSize: pageSize,
		Total:    int(count),
		Content:  resultList,
	})
}
