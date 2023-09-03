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

// @Summary 提交帖子
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
		user := getUserInfo(c)
		commit.UserInfo = user
		commit.BbsContent.UserID = user.ID
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
func getUserInfo(c *gin.Context) user.User {
	userInfo, exist := c.Get("userInfo")
	if exist == false {
		log.Error("bbs 帖子提交，用户信息不存在")
		common.FailedReturn(c, "获取用户信息失败")
		return user.User{}
	}
	return userInfo.(user.User)
}

// @Summary 查询全部贴子
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
		pageN, _ := strconv.Atoi(pageNo)
		pageS, _ := strconv.Atoi(pageSize)
		if date != "" {

			intValue, _ := strconv.Atoi(date)
			time := time.Unix(int64(intValue), 0)
			year, month, day := time.Date()
			var dateStart = fmt.Sprintf("%d-%d-%d 00:00:00", year, int(month), day)
			var dateEnd = fmt.Sprintf("%d-%d-%d 23:59:59", year, int(month), day)

			var total int64
			mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Count(&total)
			var resultList []creeper.Content
			mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Offset((pageN - 1) * pageS).Limit(pageS).Find(&resultList)

			c.JSON(http.StatusOK, common.Success(common.PageCL{
				PageNo:   pageN,
				PageSize: pageS,
				Total:    int(total),
				Content:  resultList,
			}))
		} else {
			year, month, day := time.Now().Date()
			var dateStart = fmt.Sprintf("%d-%d-%d 00:00:00", year, int(month), day)
			var dateEnd = fmt.Sprintf("%d-%d-%d 23:59:59", year, int(month), day)
			var total int64
			mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Count(&total)
			var resultList []creeper.Content
			if total <= 0 {
				mysql.DB.Model(&creeper.Content{}).Order("created_at desc").Offset((pageN - 1) * pageS).Limit(pageS).Find(&resultList)
			} else {
				mysql.DB.Model(&creeper.Content{}).Where("created_at BETWEEN ? AND ? ", dateStart, dateEnd).Order("created_at desc").Offset((pageN - 1) * pageS).Limit(pageS).Find(&resultList)
			}

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
