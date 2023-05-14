package bbs

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"jingcai/creeper"
	ilog "jingcai/log"
	"jingcai/mysql"
	"jingcai/user"
	"net/http"
	"strconv"
	"time"
)

type BBS struct {
	//用户信息
	UserInfo user.User

	//论坛信息
	BbsContent creeper.Content
}

var log = ilog.Logger

// @Description 提交论坛的帖子
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body bbs.BBS false "提交对象"
// @Router /bbs/commit [post]
func CommitHandler(c *gin.Context) {
	var commit BBS
	c.Header("Content-Type", "application/json; charset=utf-8")
	err := c.BindJSON(&commit)
	if err == nil {
		log.Info("commit ====> ", commit)
		//TODO save user info
		commit.BbsContent.UserID = 0
		commit.BbsContent.CreatedAt = time.Now()
		commit.BbsContent.UpdatedAt = time.Now()

		mysql.DB.AutoMigrate(&creeper.Content{})
		mysql.DB.Create(&commit.BbsContent)
		c.JSON(http.StatusOK, common.Success(commit))
	} else {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, &common.BaseResponse{
			Code:    0,
			Message: "参数获取失败",
		})
	}

}

// @Description 提交论坛的帖子
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param date query   uint false "日期 unix time"
// @param pageNo query int   true  "页码"
// @param pageSize query int  true  "每页条数"
// @Router /bbs/list [get]
func ListHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	pageNo := c.Query("pageNo")
	pageSize := c.Query("pageSize")
	date := c.Query("date")

	if pageNo != "" && pageSize != "" {
		var content creeper.Content
		pageN, _ := strconv.Atoi(pageNo)
		pageS, _ := strconv.Atoi(pageSize)
		if date != "" {

			intValue, _ := strconv.Atoi(date)
			time := time.Unix(int64(intValue), 0)
			year, month, day := time.Date()
			var dateStart = fmt.Sprintf("%s-%s-%s 00:00:00", year, month, day)
			var dateEnd = fmt.Sprintf("%s-%s-%s 23:59:59", year, month, day)

			var total int64
			mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Count(&total)
			var resultList []creeper.Content
			mysql.DB.Select(&content, "created_at BETWEEN ? AND ?  limit ?,?", dateStart, dateEnd, (pageN-1)*pageS, pageN*pageS).Find(&resultList)

			c.JSON(http.StatusOK, common.Success(common.PageCL{
				PageNo:   pageN,
				PageSize: pageS,
				Total:    int(total),
				Content:  resultList,
			}))
		} else {
			year, month, day := time.Now().Date()
			var dateStart = fmt.Sprintf("%s-%s-%s 00:00:00", year, int(month), day)
			var dateEnd = fmt.Sprintf("%s-%s-%s 23:59:59", year, int(month), day)
			var total int64
			mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Count(&total)
			var resultList []creeper.Content
			mysql.DB.Select(&content, "created_at BETWEEN ? AND ?  limit ?,?", dateStart, dateEnd, (pageN-1)*pageS, pageN*pageS).Find(&resultList)
			c.JSON(http.StatusOK, common.Success(common.PageCL{
				PageNo:   pageN,
				PageSize: pageS,
				Total:    int(total),
				Content:  resultList,
			}))
		}

	} else {
		c.JSON(http.StatusInternalServerError, &common.BaseResponse{
			Code:    0,
			Message: "分页参数获取失败",
		})
	}
}
