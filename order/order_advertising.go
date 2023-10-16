package order

import (
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"time"
)

type WinUserPO struct {
	Phone string
	//用户昵称
	Name string
	//金额
	Bonus float32
	//中奖时间
	Time time.Time
	//订单号
	OrderId string
	UserId  uint
	//头像
	Avatar string
}

// @Summary 中奖名单 接口
// @Description 中奖名单
// @Tags advertising 广告
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/advertising/win-list [post]
func GetWinUserList(c *gin.Context) {
	var orders []Order
	mysql.DB.Where("win", true).Order("created_at desc").Limit(50).Find(&orders)
	var winpos []WinUserPO = make([]WinUserPO, 0)
	var userIds []uint
	for _, order := range orders {
		winpos = append(winpos, WinUserPO{
			Bonus:   order.Bonus,
			Time:    order.CreatedAt,
			OrderId: order.UUID,
			UserId:  order.UserID,
		})
		userIds = append(userIds, order.UserID)
	}

	mapper := user.FindUsserMapById(userIds)

	for i := 0; i < len(winpos); i++ {
		uv := mapper[winpos[i].UserId]
		winpos[i].Phone = uv.Phone
		winpos[i].Name = uv.Name
		winpos[i].Avatar = uv.Avatar
	}
	common.SuccessReturn(c, winpos)
}
