package common

import (
	_ "crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"net/http"
	"strconv"
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
	c.Abort()
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
// @Router /api/salt [get]
func Salt(c *gin.Context) {
	privateKey, pubKey := GenRsaKey()
	decodePubKey := base64.StdEncoding.EncodeToString([]byte(pubKey))
	CacheJingCai.Add(decodePubKey, SALT_OUT_TIME, privateKey)
	SuccessReturn(c, decodePubKey)
}
func getNum(num int) string {

	if num < 10 {
		return fmt.Sprintf("0%d", num)
	} else {
		return strconv.Itoa(num)
	}

}
func GetMatchFinishedTime(time2 time.Time) time.Time {
	now := time.Now()
	var dateEnd string
	if now.Weekday() == 0 || now.Weekday() == 6 {
		dateEnd = fmt.Sprintf("%d-%s-%s 22:55:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))

	} else {
		dateEnd = fmt.Sprintf("%d-%s-%s 21:55:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
	}

	time1, err := time.ParseInLocation("2006-01-02 15:04:05", dateEnd, time.Local)

	if err != nil {
		fmt.Println(err)
	}
	if time1.UnixMicro() > time2.UnixMicro() {
		timeBefore := time2.UnixMicro() - 300000
		return time.UnixMicro(timeBefore).Local()
	} else {
		return time1
	}

}

func GetDateStartAndEnd(time time.Time) (string, string) {
	now := time
	var dateEnd string
	if now.Weekday() == 0 || now.Weekday() == 6 {
		dateEnd = fmt.Sprintf("%d-%s-%s 22:55:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))

	} else {
		dateEnd = fmt.Sprintf("%d-%s-%s 21:55:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
	}

	var dateStart string
	dateStart = fmt.Sprintf("%d-%s-%s 00:00:00", now.Year(), getNum(int(now.Month())), getNum(int(now.Day())))
	return dateStart, dateEnd
}

/*func IsEmpty(obj interface{}) bool {
	value := reflect.ValueOf(obj).Elem() // 获取指针对应的值
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		return false // 如果是非nil指针类型则不为空
	} else if value.NumField() > 0 {
		for i := 0; i < value.NumField(); i++ {
			field := value.Type().Field(i)
			// 只有当字段没有被标记为omitempty时才进行判断
			if field.Tag != "omitempty" && !isZero(value.Field(i)) {
				return false // 存在非空字段则不为空
			}
		}
	}

	return true
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Slice, reflect.Array:
		return v.IsNil()
	default:
		panic("unsupported type")
	}
}*/
