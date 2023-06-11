package common

import (
	_ "crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"net/http"
	"time"
)

var CacheJingCai = cache2go.Cache("jingcai")

const SALT_OUT_TIME = 10 * time.Minute

// BaseResponse 返回对象
type BaseResponse struct {
	//1 成功 0 失败
	Code int `json:"code"`

	//错误信息
	Message string `json:"message"`

	// in: body
	Content interface{} `json:"content"`
}

// PageCL 分页
type PageCL struct {
	//页码
	PageNo int
	//每页大小
	PageSize int
	//总条数
	Total int
	//内容
	Content interface{}
}

func Success(c interface{}) *BaseResponse {
	return &BaseResponse{
		Code:    1,
		Message: "执行成功",
		Content: c,
	}
}

func Failed() *BaseResponse {
	return &BaseResponse{
		Code:    0,
		Message: "执行失败",
	}
}

func SuccessReturn(c *gin.Context, content interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, &BaseResponse{
		Code:    1,
		Message: "执行成功",
		Content: content,
	})
}

func FailedReturn(c *gin.Context, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusInternalServerError, &BaseResponse{
		Code:    0,
		Message: message,
	})
}

func FailedAuthReturn(c *gin.Context, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusUnauthorized, &BaseResponse{
		Code:    0,
		Message: message,
	})
}

// @Summary 获取加密的公钥
// @Description 公钥 默认10分钟过期
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /salt [get]
func Salt(c *gin.Context) {
	privateKey, pubKey := GenRsaKey()
	decodePubKey := base64.StdEncoding.EncodeToString([]byte(pubKey))
	CacheJingCai.Add(decodePubKey, SALT_OUT_TIME, privateKey)
	SuccessReturn(c, decodePubKey)
}

func GetMatchFinishedTime(time2 time.Time) time.Time {
	now := time.Now()
	var dateEnd string
	if now.Weekday() == 0 || now.Weekday() == 6 {
		dateEnd = fmt.Sprintf("%d-%d-%d 22:55:00", now.Year(), now.Month(), now.Day())
	} else {
		dateEnd = fmt.Sprintf("%d-%d-%d 21:55:00", now.Year(), now.Month(), now.Day())
	}

	time, _ := time.ParseInLocation("2006-01-02 15:04:05", dateEnd, time.Local)
	if time.UnixMicro() > time2.UnixMicro() {
		return time2
	} else {
		return time
	}

}
