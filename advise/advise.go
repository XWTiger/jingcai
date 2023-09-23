package advise

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/mysql"
	"time"
)

var log = ilog.Logger

type NotificationPO struct {
	gorm.Model
	//通知内容，富文本
	Content string `minLength:"2"`

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
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body Notification false "通告对象"
// @Router /api/super/notify [post]
func Create(c *gin.Context) {

	var notify Notification
	c.BindJSON(&notify)
	mysql.DB.AutoMigrate(&NotificationPO{})
	if notify != (Notification{}) {
		timeExpire, _ := time.ParseInLocation("2006-01-02 15:04:05", notify.Expired, time.Local)
		if err := mysql.DB.Create(&NotificationPO{Content: notify.Content, Expired: timeExpire}).Error; err != nil {
			common.FailedReturn(c, "创建失败")
		}
	}
}

// @Summary 查询通告
// @Description 查询通告
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
		common.FailedReturn(c, "当前没有通告")
		return
	}
	common.SuccessReturn(c, notify)

}
