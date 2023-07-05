package cache

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	uuid "github.com/satori/go.uuid"
	"io"
	"jingcai/common"
	ilog "jingcai/log"
	"net/http"
	"strings"
	"time"
)

var lotteryCahe = cache2go.Cache("lottery")

const TOKEN_TIME_OUT = 8 * time.Hour
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
		//log.Info(string(bytes))
		lotteryCahe.Add(uuid, TOKEN_TIME_OUT, nil)
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

func GetOnTimeFootballMatch(uuid string) *FootBallGames {
	if !lotteryCahe.Exists(uuid) {
		return nil
	}
	var url = "http://127.0.0.1:8090/lottery/sports/jc/mixed"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result FootBallGames
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("转换足彩后台查询实时对象失败")
		return nil
	}
	lotteryCahe.Add(uuid, 10*time.Minute, body)
	return &result
}
