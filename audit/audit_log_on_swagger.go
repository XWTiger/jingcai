package audit

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/cache2go"
	"github.com/swaggo/swag"
	"gorm.io/gorm"
	"jingcai/common"
	"jingcai/mysql"
	"jingcai/user"
	"regexp"
	"strings"
)

var userCahe = cache2go.Cache("user")

type AuditLog struct {
	gorm.Model
	//用户名称
	UserName string

	//url
	Path string
	//接口描述
	Summery string
	//接口方式 GET PUT POST DELETE
	Method string
	//执行状态
	Status bool

	Ip string
	//手机
	Phone string
}

type urlInfo struct {
	//url
	Path string
	//接口描述
	Summery string
}

var fullPathMapper = make(map[string]*urlInfo, 0)
var pathVariableMapper = make(map[string]*urlInfo, 0)

// 根据swagger json 设置日志配置
func InitAudit() {

	/*dirs, _ := os.ReadDir("../")
	for _, dir := range dirs {
		fmt.Println(dir.Name())
		fmt.Println(dir.Info())
		fmt.Println(dir.Type())
	}*/
	var obj map[string]interface{}
	cont, _ := swag.ReadDoc("swagger")
	json.Unmarshal([]byte(cont), &obj)
	fmt.Println("===========================")
	val := obj["paths"].(map[string]interface{})
	for k, v := range val {
		fmt.Println(k)
		var url = urlInfo{}
		var parttern string
		//path variable
		if strings.Contains(k, ":") {
			arrK := strings.Split(k, ":")

			str := strings.ReplaceAll(arrK[0], "/", "\\/")
			for i := 1; i < len(arrK); i++ {

				index := strings.Index(arrK[i], "/")
				if index > 0 {
					real := "\\w*" + arrK[i][0:index]
					str += real
				} else {
					str += "\\w*"
				}
			}
			parttern = str
		}
		url.Path = k
		if strings.Contains(k, ":") {
			pathVariableMapper[parttern] = &url
		} else {
			fullPathMapper[k] = &url
		}
		valin := v.(map[string]interface{})
		if valin["get"] != nil {
			detail := valin["get"].(map[string]interface{})
			fmt.Println(detail["summary"])
			if detail["summary"] == nil {
				if detail["description"] != nil {
					url.Summery = detail["description"].(string)
				} else {
					url.Summery = k
				}
				continue
			}
			url.Summery = detail["summary"].(string)
			//summary
			continue

		}
		if valin["post"] != nil {
			detail := valin["post"].(map[string]interface{})
			fmt.Println(detail["summary"])
			if detail["summary"] == nil {
				if detail["description"] != nil {
					url.Summery = detail["description"].(string)
				} else {
					url.Summery = k
				}
				continue
			}
			url.Summery = detail["summary"].(string)
			//summary
			continue

		}

		if valin["put"] != nil {
			detail := valin["put"].(map[string]interface{})
			fmt.Println(detail["summary"])
			if detail["summary"] == nil {
				if detail["description"] != nil {
					url.Summery = detail["description"].(string)
				} else {
					url.Summery = k
				}
				continue
			}
			url.Summery = detail["summary"].(string)
			//summary
			continue

		}

		if valin["delete"] != nil {
			detail := valin["delete"].(map[string]interface{})
			fmt.Println(detail["summary"])
			if detail["summary"] == nil {
				if detail["description"] != nil {
					url.Summery = detail["description"].(string)
				} else {
					url.Summery = k
				}
				continue
			}
			url.Summery = detail["summary"].(string)
			//summary
			continue

		}
	}
	return
}
func getSumery(path string) string {
	for s, info := range pathVariableMapper {
		ok, _ := regexp.MatchString(s, path)
		if ok {
			return info.Summery
		}
	}
	info := fullPathMapper[path]
	if common.IsEmpty(info) {
		return ""
	}
	return info.Summery
}

func AuditHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var audit = AuditLog{}
		var token string
		token = c.Query("token") // 访问令牌
		if token == "" {
			token = c.GetHeader("token")
			if token == "" {
				fmt.Println("== get token failed ===")
				audit.Path = c.FullPath()
				audit.Summery = getSumery(c.FullPath())
				audit.Status = (c.Writer.Status() == 200)
				audit.UserName = "过客"
				audit.Ip = c.RemoteIP()
				audit.Method = c.Request.Method
				mysql.DB.Save(&audit)
				return
			}
		}
		res, err := userCahe.Value(token)
		if err != nil {
			audit.Path = c.FullPath()
			audit.Summery = getSumery(c.FullPath())
			audit.Status = (c.Writer.Status() == 200)
			audit.UserName = "过客"
			audit.Ip = c.RemoteIP()
			audit.Method = c.Request.Method
		} else {
			// 验证通过，会继续访问下一个中间件
			user := res.Data().(user.User)
			audit.Path = c.FullPath()
			audit.Summery = getSumery(c.FullPath())
			audit.Status = (c.Writer.Status() == 200)
			audit.UserName = user.Name
			audit.Ip = c.RemoteIP()
			audit.Method = c.Request.Method
			audit.Phone = user.Phone
		}
		mysql.DB.Save(&audit)
		c.Next()
	}
}
