package cache

import (
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	uuid "github.com/satori/go.uuid"
	"io"
	"jingcai/common"
	ilog "jingcai/log"
	"strings"
	"time"
)

var lotteryCahe = cache2go.Cache("lottery")

const TOKEN_TIME_OUT = 1 * time.Hour
const BIND = "[::1]"
const Local = "127.0.0.1"

var log = ilog.Logger

// @Summary 内部缓存使用接口
// @Description 内部缓存使用接口
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @param param body interface{} false "提交对象"
// @Router /cache [post]
func Set(c *gin.Context) {
	log.Info("remote ip: ", c.Request.RemoteAddr)
	if strings.HasPrefix(c.Request.RemoteAddr, BIND) || strings.HasPrefix(c.Request.RemoteAddr, Local) {
		uuid := uuid.NewV4().String()
		bytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error("read body failed: ", err)
			common.FailedReturn(c, "获取body失败")
			return
		}
		if len(bytes) <= 0 {
			log.Error("body is empty")
			common.FailedReturn(c, "获取body失败")
			return
		}
		//log.Info(string(bytes))
		lotteryCahe.Add(uuid, TOKEN_TIME_OUT, string(bytes))
		common.SuccessReturn(c, uuid)

	} else {
		common.FailedReturn(c, "该接口不对外")
		return
	}

}

func Remove(key string) {
	lotteryCahe.Delete(key)
}

func Get(key string) interface{} {

	item, err := lotteryCahe.Value(key)
	if err != nil {
		return nil
	}
	return item.Data()
}
