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
// @Router /api/order [get]
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

	var footViews = make([]FootView, 0)
	if err := mysql.DB.Model(&FootView{}).Where("match_id=?", matchId).Find(&footViews).Error; err != nil {
		log.Error("根据比赛id 查比赛详情失败")
		return nil, err
	}
	var mapper = make(map[uint][]FootView)
	var ids = make([]uint, 0)
	for i := 0; i < len(footViews); i++ {
		_, ok := mapper[footViews[i].BetId]
		if ok {
			mapper[footViews[i].BetId] = append(mapper[footViews[i].BetId], footViews[i])
		} else {
			var list = make([]FootView, 0)
			list = append(list, footViews[i])
			mapper[footViews[i].BetId] = list
		}
		ids = append(ids, footViews[i].BetId)
	}

	var betResult = make([]Bet, 0)
	if err := mysql.DB.Debug().Model(Bet{}).Where("id In ?", ids).Find(&betResult).Error; err != nil {
		log.Error("查询bet 失败！")
		log.Error(err)
		return nil, err
	}

	for index, bet := range betResult {
		betResult[index].Group = mapper[bet.ID]
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
