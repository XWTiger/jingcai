package order

import (
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"sort"
	"strconv"
	"strings"
)

type OrderImage struct {
	gorm.Model
	//访问地址
	Url string

	//order uuid
	ParentId string
}

func getImageByOrderId(oid string) []OrderImage {

	var images = make([]OrderImage, 0)
	if err := mysql.DB.Model(&OrderImage{}).Where(&OrderImage{ParentId: oid}).Find(&images).Error; err != nil {
		log.Error(err)
		return images
	}
	return images
}

type UploadBet struct {
	//订单id
	OrderId string `validate:"required"`

	//图片地址
	Url []string

	//如果OddChange 为true 才传该对象
	MatchOdds []MatchOdd
}

type MatchOdd struct {
	MatchId string `validate:"required"`
	//足球类型 枚举：SFP（胜负平）、BF（比分）、ZJQ(总进球)、BQSFP（半全场胜负平）
	//篮球类型 枚举：HDC （胜负）、 HILO（大小分）、 MNL（让分胜负）、 WNM（胜分差）
	Type string //`validate:"required"`
	//让球 胜平负才有，篮球就是让分
	GoalLine string
	//=================足球=========================
	//比分， 类型BF才有 s00s00 s05s02
	//半全场胜平负， 类型BQSFP  aa hh
	//总进球数， 类型ZJQ s0 - s7
	//胜负平， 类型SFP hada主负 hadd主平 hadh 主胜  hhada客负 hhadd客平 hhadh 客胜
	//=================篮球=========================
	//让分胜负， 类型HDC a 负，  h 胜
	//大小分，类型HILO l 小， h 大
	//胜负，类型MNL a 主负， h 主胜
	//胜分差，类型WNM l1 客胜1-5分  l2 6-10分 ... l6 26+分， w1 主胜1-5分 ... w6 26+分
	ScoreVsScore string `validate:"required"`
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
		log.Error(err)
		common.FailedReturn(c, "参数错误")
		return
	}
	tx := mysql.DB.Begin()

	//调整赔率
	if len(betImgObj.MatchOdds) > 0 {
		var matchs = betImgObj.MatchOdds

		for i := 0; i < len(matchs); i++ {
			var matchPo Match
			if err := tx.Model(Match{}).Where(&Match{
				MatchId: matchs[i].MatchId,
				OrderId: betImgObj.OrderId,
			}).Find(&matchPo).Error; err != nil {
				log.Error(err)
				log.Error("查不到订单比赛记录")
				common.FailedReturn(c, "查不到比赛记录！！")
				tx.Rollback()
				return
			}
			var lottery []LotteryDetail
			if lerr := tx.Model(LotteryDetail{}).Where(&LotteryDetail{ParentId: matchPo.ID}).Find(&lottery).Error; lerr != nil {
				log.Error("更新赔率 没有查询到票的信息")
				common.FailedReturn(c, "更新赔率 没有查询到票的信息！！")
				tx.Rollback()
				return
			}
			for i2, detail := range lottery {
				if matchs[i].Type == detail.Type && matchs[i].ScoreVsScore == detail.ScoreVsScore {
					lottery[i2].Odds = matchs[i].Odds
					tx.Save(&lottery[i2])
				}
			}

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
			ParentId: betImgObj.OrderId,
		})
	}
	order := FindById(betImgObj.OrderId, false)
	//更新订单为已上传
	if order.BetUpload == 2 {
		//如果订单已经传了图片 那么就要删除以前的
		derr := tx.Model(&OrderImage{ParentId: betImgObj.OrderId}).Where("parent_id=?", betImgObj.OrderId).Delete(&OrderImage{ParentId: betImgObj.OrderId}).Error
		if derr != nil {
			log.Error(derr)
			common.FailedReturn(c, "图标保存失败")
			tx.Rollback()
			return
		}

	}
	if imageErr := tx.Create(&images).Error; imageErr != nil {
		common.FailedReturn(c, "图标保存失败")
		tx.Rollback()
		return
	}
	if !(order.BetUpload == 2) {
		dbErr := tx.Model(&Order{UUID: betImgObj.OrderId}).Update("bet_upload", 2).Error
		if dbErr != nil {
			log.Error(dbErr)
			common.FailedReturn(c, "更新图标失败")
			tx.Rollback()
			return
		}
	}
	tx.Commit()
}

// @Summary 调整足球篮球赔率接口
// @Description 调整足球篮球赔率接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param bets  body UploadBet true "管理员提交票对象"
// @Router /api/super/matches/odds [post]
func UpdateOddHandler(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	if strings.Compare(userInfo.Role, user.ADMIN) != 0 {
		common.FailedReturn(c, "该接口只提供给管理员")
		return
	}
	var betImgObj UploadBet
	err := c.BindJSON(&betImgObj)
	if err != nil {
		log.Error(err)
		common.FailedReturn(c, "参数错误")
		return
	}

	//调整赔率
	if len(betImgObj.MatchOdds) > 0 {
		tx := mysql.DB.Begin()
		var matchs = betImgObj.MatchOdds

		for i := 0; i < len(matchs); i++ {
			var matchPo Match
			if err := tx.Model(&Match{}).Where(&Match{
				MatchId: matchs[i].MatchId,
				OrderId: betImgObj.OrderId,
			}).Find(&matchPo).Error; err != nil {
				log.Error(err)
				log.Error("查不到订单比赛记录")
				common.FailedReturn(c, "查不到比赛记录！！")
				tx.Rollback()
				return
			}
			var lottery []LotteryDetail
			if lerr := tx.Model(LotteryDetail{}).Where(&LotteryDetail{ParentId: matchPo.ID}).Find(&lottery).Error; lerr != nil {
				log.Error("更新赔率 没有查询到票的信息")
				common.FailedReturn(c, "更新赔率 没有查询到票的信息！！")
				tx.Rollback()
				return
			}
			for i2, detail := range lottery {
				if matchs[i].Type == detail.Type && matchs[i].ScoreVsScore == detail.ScoreVsScore {
					lottery[i2].Odds = matchs[i].Odds
					tx.Save(&lottery[i2])
				}
			}
			bets, err := getBetByMatchIdAndOid(matchs[i].MatchId, betImgObj.OrderId)
			if err == nil {
				for j, bet := range bets {
					for i3, view := range bets[j].Group {
						if view.MatchId == matchs[i].MatchId {
							bet.Group[i3].Odd = matchs[i].Odds
							err := tx.Model(&FootView{Model: gorm.Model{
								ID: view.ID,
							}}).Update("odd", bet.Group[i3].Odd).Error
							if err != nil {
								log.Error(err)
							}
							break
						}
					}
					dbonus := decimal.NewFromInt(2)
					for _, view := range bet.Group {
						dbonus = dbonus.Mul(decimal.NewFromFloat32(view.Odd))
					}
					value, _ := dbonus.Float64()
					bet.Bonus = float32(value)
					tx.Save(&bet)
				}
			}

		}
		tx.Commit()
	}

	derr := mysql.DB.Model(&Order{UUID: betImgObj.OrderId}).Update("bet_upload", 1).Error
	if derr != nil {
		common.FailedReturn(c, "更新传票状态失败")
		return
	}
	common.SuccessReturn(c, "执行成功")
}

// @Summary 管理员订单查询接口
// @Description 订单查询接口
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param lotteryType  query string false "足彩（FOOTBALL） 大乐透（SUPER_LOTTO）  排列三（P3） 篮球(BASKETBALL) 七星彩（SEVEN_STAR） 排列五（P5）"
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @param hasImage query string false "只看未上传图片， 0代表 否 1代表是"
// @param startDate  query string false "截止时间 2023-01-01 21:27:00"
// @param endDate  query string false "截止时间 2023-01-01 21:27:00"
// @param deadTimeSort query bool false "true or false"
// @Router /api/super/order [get]
func AdminOrderList(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)

	lotteryType := c.Query("lotteryType")
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	hasImage := c.Query("hasImage")
	deadTimeSort := c.Query("deadTimeSort")

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// TODO
	if strings.Compare(userInfo.Role, user.ADMIN) != 0 {
		common.FailedReturn(c, "该接口只提供给管理员")
		return
	}
	var param = Order{}

	if lotteryType != "" {
		param.LotteryType = lotteryType
	}

	var list = make([]Order, 0)
	var resultList = make([]*OrderVO, 0)
	var query *gorm.DB
	var query2 *gorm.DB

	if lotteryType != "" {
		log.Info("======1=====")
		query = mysql.DB.Debug().Model(&param).Select("orders.*").Where("all_win_id > 0 and pay_status = true  and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')").Joins("Left JOIN all_wins on orders.all_win_id = all_wins.id and all_wins.`parent_id`=0  and all_wins.`status` = true" + " and orders.lottery_type='" + lotteryType + "'")
		query2 = mysql.DB.Debug().Model(&param).Select("orders.* ").Where("orders.deleted_at is null and pay_status = true and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')" + " and orders.lottery_type='" + lotteryType + "'")
	} else {
		query = mysql.DB.Debug().Model(&param).Select("orders.*").Where("all_win_id > 0 and pay_status = true  and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')").Joins("Left JOIN all_wins on orders.all_win_id = all_wins.id and all_wins.`parent_id`=0  and all_wins.`status` = true")
		query2 = mysql.DB.Debug().Model(&param).Select("orders.* ").Where("orders.deleted_at is null and pay_status = true and (orders.save_type='TOMASTER' or  orders.save_type='ALLWIN')")
	}
	var count int64
	if hasImage == "1" {
		query.Where("orders.bet_upload != 2")
		query2.Where("orders.bet_upload != 2")
	}
	if startDate != "" && endDate != "" {
		query.Where("dead_time between " + startDate + " and " + endDate)
		query2.Where("dead_time between " + startDate + " and " + endDate)
	}
	var orderStr = "orders.created_at desc"
	if deadTimeSort == "true" {
		log.Info("======2=====")
		orderStr = "orders.dead_time asc"
	}
	var listF []Order
	var listS []Order
	query.Order(orderStr).Find(&listF)
	query2.Order(orderStr).Find(&listS)
	for i := 0; i < len(listF); i++ {
		list = append(list, listF[i])
	}
	for i := 0; i < len(listS); i++ {
		list = append(list, listS[i])
	}

	count = int64(len(listF) + len(listS))
	//mysql.DB.Raw("? union ? ", query, query2).Count(&count)
	//mysql.DB.Raw("? union ? ", query, query2).Order(orderStr).Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)
	if len(list) <= 0 {
		common.SuccessReturn(c, &common.PageCL{
			PageNo:   page,
			PageSize: pageSize,
			Total:    int(count),
			Content:  resultList,
		})
		return
	}
	if deadTimeSort == "true" {
		sort.Slice(list, func(i, j int) bool {
			return list[j].DeadTime.UnixMicro() < list[i].CreatedAt.UnixMicro()
		})
	} else {
		sort.Slice(list, func(i, j int) bool {
			return list[i].CreatedAt.UnixMicro() > list[j].CreatedAt.UnixMicro()
		})
	}

	start := (page - 1) * pageSize
	var end int
	if (start + pageSize) > len(list) {
		end = len(list)
	} else {
		end = start + pageSize
	}

	for index, order := range list[start:end] {
		var mathParam = Match{
			OrderId: list[start+index].UUID,
		}
		if strings.Compare(order.LotteryType, FOOTBALL) == 0 || strings.Compare(order.LotteryType, BASKETBALL) == 0 {
			var matchList = make([]Match, 0)
			mysql.DB.Model(&mathParam).Where("order_id=?", list[start+index].UUID).Find(&matchList)
			list[start+index].Matches = matchList
			for i, match := range matchList {
				var detailParam = LotteryDetail{
					ParentId: match.ID,
				}
				var detailList = make([]LotteryDetail, 0)
				mysql.DB.Model(&detailParam).Where("parent_id=?", match.ID).Find(&detailList)
				matchList[i].Combines = detailList
			}
		}

		resultList = append(resultList, &OrderVO{
			Order:  &list[start+index],
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
