package admin

import (
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"strconv"
)

// @Summary 查看投诉
// @Description 查看投诉
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @Router /api/super/complains [get]
func ListComplain(c *gin.Context) {
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	pg, err := strconv.Atoi(pageNo)
	pgsize, err := strconv.Atoi(pageSize)
	if err != nil {
		common.FailedReturn(c, "分页参数获取失败")
		return
	}
	var complains []user.Complain
	var count int64
	mysql.DB.Debug().Model(&user.Complain{}).Count(&count).Offset((pg - 1) * pgsize).Limit(pgsize).Find(&complains)
	common.SuccessReturn(c, common.PageCL{
		PageNo:   pg,
		PageSize: pgsize,
		Total:    int(count),
		Content:  complains,
	})
}
