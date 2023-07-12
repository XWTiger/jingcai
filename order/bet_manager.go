package order

import (
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
)

// @Summary 查询足彩组合详情
// @Description 查询足彩组合详情
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param orderId  query string false "订单号"
// @Router /order [get]
func GetBetByOrder(c *gin.Context) {

	var user = user.FetUserInfo(c)
	orderId := c.Query("orderId")
	var params = Bet{
		OrderId: orderId,
		UserId:  user.ID,
	}
	var betResult = make([]Bet, 0)
	if err := mysql.DB.Model(&params).Find(&betResult).Error; err != nil {
		log.Error(err)
		common.SuccessReturn(c, betResult)
		return
	}
	for _, bet := range betResult {
		var fparam = FootView{
			BetId: bet.ID,
		}
		var footViews = make([]FootView, 0)
		mysql.DB.Model(&fparam).Find(&footViews)
		bet.Group = footViews
	}
	common.SuccessReturn(c, betResult)
}

func getBetByMatchId(matchId string) ([]Bet, error) {
	var params = Bet{
		MatchId: matchId,
	}
	var betResult = make([]Bet, 0)
	if err := mysql.DB.Debug().Model(&params).Where("match_id = ?", matchId).Find(&betResult).Error; err != nil {
		log.Error("查询bet 失败！")
		log.Error(err)
		return nil, err
	}
	for index, bet := range betResult {
		var fparam = FootView{
			BetId: bet.ID,
		}
		var footViews = make([]FootView, 0)
		mysql.DB.Model(&fparam).Where("bet_id=?", bet.ID).Find(&footViews)
		betResult[index].Group = footViews
	}
	return betResult, nil
}

func getBetByOrderId(orderId string) []Bet {
	var bets = make([]Bet, 0)
	mysql.DB.Debug().Model(&Bet{
		OrderId: orderId,
	}).Where("order_id=?", orderId).Find(&bets)
	return bets
}
