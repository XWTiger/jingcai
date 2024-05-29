package advise

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"jingcai/user"
	"strconv"
	"time"
)

var log = ilog.Logger

type NotificationPO struct {
	gorm.Model
	//通知内容，富文本
	Content string `minLength:"2"`

	//店主id
	UserId uint

	//过期时间
	Expired time.Time
}

type Notification struct {
	gorm.Model
	//通知内容，富文本
	Content string `minLength:"2"`

	//过期时间 2023-05-24 17:01:11
	Expired string
}

// @Summary 创建通告
// @Description 创建通告
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Notification false "通告对象"
// @Router /api/super/notify [post]
func Create(c *gin.Context) {
	var userInfo = user.FetUserInfo(c)
	var notify Notification
	c.BindJSON(&notify)
	mysql.DB.AutoMigrate(&NotificationPO{})
	if notify != (Notification{}) {
		timeExpire, _ := time.ParseInLocation("2006-01-02 15:04:05", notify.Expired, time.Local)
		if err := mysql.DB.Create(&NotificationPO{Content: notify.Content, Expired: timeExpire, UserId: userInfo.ID}).Error; err != nil {
			common.FailedReturn(c, "创建失败")
		}
	}
}

// @Summary 通告列表
// @Description 通告列表
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param pageNo  query int true "页码"
// @param pageSize  query int true "每页大小"
// @Router /api/super/notify/list [get]
func List(c *gin.Context) {
	time := time.Now()
	date := time.Format("2006-01-02 15:04:05")
	var notify []NotificationPO
	page, _ := strconv.Atoi(c.Query("pageNo"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	c.BindJSON(&notify)
	var count int64
	mysql.DB.AutoMigrate(&NotificationPO{})
	if err := mysql.DB.Model(&NotificationPO{}).Where("expired  >= ? ", date).Count(&count).Offset((page - 1) * pageSize).Limit(pageSize).Find(&notify).Error; err != nil {
		log.Error(err)
		common.SuccessReturn(c, "欢迎来到黑马门店助手！")
		return
	}
	common.SuccessReturn(c, common.PageCL{
		PageNo:   page,
		PageSize: pageSize,
		Total:    int(count),
		Content:  notify,
	})
}

// @Summary 删除通告
// @Description 删除通告
// @Tags owner 店主
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/super/notify/{id} [delete]
func Delete(c *gin.Context) {
	id := c.Param("id")
	rid, _ := strconv.Atoi(id)
	if err := mysql.DB.Model(&Notification{Model: gorm.Model{ID: uint(rid)}}).Delete(&Notification{Model: gorm.Model{ID: uint(rid)}}).Error; err != nil {
		log.Error(err)
		return
	}
	common.SuccessReturn(c, "删除成功")
}

// @Summary 查询通告
// @Description 查询通告
// @Tags notify 通告
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/notify [get]
func Query(c *gin.Context) {
	time := time.Now()
	date := time.Format("2006-01-02 15:04:05")
	var notify NotificationPO
	mysql.DB.AutoMigrate(&NotificationPO{})
	if err := mysql.DB.Model(&NotificationPO{}).Where("expired  >= ?", date).First(&notify).Error; err != nil {
		log.Error(err)
		common.SuccessReturn(c, "欢迎来到黑马门店助手！")
		return
	}
	common.SuccessReturn(c, notify)

}
