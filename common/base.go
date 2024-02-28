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
func GetNum(num int) string {

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
		dateEnd = fmt.Sprintf("%d-%s-%s 22:55:00", now.Year(), GetNum(int(now.Month())), GetNum(int(now.Day())))

	} else {
		dateEnd = fmt.Sprintf("%d-%s-%s 21:55:00", now.Year(), GetNum(int(now.Month())), GetNum(int(now.Day())))
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
		dateEnd = fmt.Sprintf("%d-%s-%s 22:55:00", now.Year(), GetNum(int(now.Month())), GetNum(int(now.Day())))

	} else {
		dateEnd = fmt.Sprintf("%d-%s-%s 21:55:00", now.Year(), GetNum(int(now.Month())), GetNum(int(now.Day())))
	}

	var dateStart string
	dateStart = fmt.Sprintf("%d-%s-%s 00:00:00", now.Year(), GetNum(int(now.Month())), GetNum(int(now.Day())))
	return dateStart, dateEnd
}

/*
	func IsEmpty(obj interface{}) bool {
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
	}
*/
var dict = map[string]string{
	"DIRECT":        "直选",
	"P3":            "排列3",
	"P5":            "排列5",
	"ALL_WIN":       "合买",
	"NO_BONUS":      "未中奖",
	"BONUS_READY":   "已兑奖",
	"BONUS_NO_PAY":  "未兑奖",
	"DTQQ":          "前区胆拖",
	"DTHQ":          "后区胆拖",
	"DTSQ":          "双区胆拖",
	"FSQQ":          "前区复式",
	"FSHQ":          "后区复式",
	"FSSQ":          "双区复式",
	"FSSTAR":        "按位复式",
	"NO_PAY":        "未支付",
	"SIGNAL":        "单注",
	"C3":            "组合3",
	"C6":            "组合6",
	"ZX_GSB":        "直选",
	"ZX_FS_QZH":     "直选复式全组合",
	"ZX_FS":         "直选复式",
	"C3_FS":         "组选三复式",
	"C3_DT":         "组选三胆拖",
	"C6_FS":         "组选六复式",
	"C6_DT":         "组选六胆拖",
	"TOMASTER":      "提交到店",
	"SCORE":         "积分",
	"RMB":           "人民币",
	"TEMP":          "临时保存",
	"QQ":            "前区",
	"HQ":            "后区",
	"QQD":           "前区胆",
	"QQT":           "前区拖",
	"HQD":           "后区胆",
	"HQT":           "后区拖",
	"DIRECT_PLUS":   "直选多注",
	"RANDOM":        "随机一注",
	"RANDOM_PLUS":   "随机多注",
	"D":             "胆码",
	"T":             "拖码",
	"ZX_FS_GSB":     "直选复式(按位)",
	"ZX_FS_ZH3":     "直选组合三不同",
	"C6_PLUS":       "组选",
	"ZX_FS_QZH_ET":  "直选复式二同",
	"ZX_FS_QZH_WBT": "直选复式五不同",
	"ZX_FS_3T":      "直选组合三同",
	"ZX_FS_2T":      "直选组合二同",
	"ZX_FS_DT":      "直选组合胆拖",
	"FREE_SCORE":    "赠送积分",
}

func GetDictsByKey(key string) string {
	return dict[key]
}

// @Summary 获取字典
// @Description 键值对
// @Accept json
// @Produce json
// @Success 200 {object} common.BaseResponse
// @failure 500 {object} common.BaseResponse
// @Router /api/dict [get]
func Dict(c *gin.Context) {
	SuccessReturn(c, dict)
}
