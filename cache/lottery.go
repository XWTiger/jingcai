package cache

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/muesli/cache2go"
	uuid "github.com/satori/go.uuid"
	"io"
	"jingcai/common"
	ilog "jingcai/log"
	"jingcai/lottery"
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
// @Router /api/cache [post]
func Set(c *gin.Context) {
	log.Info("remote ip: ", c.Request.RemoteAddr)
	if strings.HasPrefix(c.Request.RemoteAddr, BIND) || strings.HasPrefix(c.Request.RemoteAddr, Local) {
		uuid := uuid.NewV4().String()
		//log.Info(string(bytes))
		lotteryCahe.Add(uuid, TOKEN_TIME_OUT, nil)
		var body map[string]interface{}
		c.Bind(&body)
		for k, v := range body {
			if !lottery.LotteryStatistics.Exists(k) {
				lottery.LotteryStatistics.Add(k, 6*time.Hour, v)
			}
		}
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

func GetOnTimeFootballMatch(uuid string) (*FootBallGames, error) {
	var err error
	if !lotteryCahe.Exists(uuid) {
		err = errors.New("uuid 不存在")
		log.Error("uuid 不存在")
	} else {
		err = nil
	}
	var url = "http://127.0.0.1:8090/lottery/sports/jc/mixed"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result LotteryResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("转换足彩后台查询实时对象失败")
		return nil, err
	}
	lotteryCahe.Add(uuid, 10*time.Minute, string(body))
	return &result.Content, err
}

func GetOnTimeBasketBallMatch(uuid string) (*BasketBallGames, error) {
	var err error
	if !lotteryCahe.Exists(uuid) {
		err = errors.New("uuid 不存在")
		log.Error("uuid 不存在")
	} else {
		err = nil
	}
	var url = "http://127.0.0.1:8090/lottery/sports/basketball/jc/mixed"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result BasketBallGames
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Error("转换足彩后台查询实时对象失败")
		return nil, err
	}
	lotteryCahe.Add(uuid, 10*time.Minute, string(body))
	return &result, err
}
