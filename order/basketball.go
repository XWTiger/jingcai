package order

import (
	"github.com/gin-gonic/gin"
	"jingcai/cache"
	"jingcai/common"
	"jingcai/mysql"
)

func basketball(c *gin.Context, order *Order) {
	if len(order.Matches) <= 0 {
		common.FailedReturn(c, "比赛场数不能为空")
		return
	}
	tx := mysql.DB.Begin()

	//回填比赛信息 以及反填胜率
	officalMatch := cache.GetOnTimeFootballMatch(order.LotteryUuid)
	if officalMatch == nil {
		common.FailedReturn(c, "查公布信息异常， 请联系管理员！")
		return
	}
	fillStatus := fillMatches(*officalMatch, order, c, tx)
	if fillStatus == nil {
		return
	}

	//保存所有组合
	common.SuccessReturn(c, order.UUID)
}
