package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"jingcai/common"
	"strings"
)

var WHITE_LIST = [...]string{"sss", "sdsaf"}

const ROLE_USER = "User"
const ROLE_ADMIN = "Admin"
const USER_INFO = "userInfo"

func Authorize() gin.HandlerFunc {
	return func(c *gin.Context) {

		//白名单

		var token string
		token = c.Query("token") // 访问令牌
		if token == "" {
			token = c.GetHeader("token")
			if token == "" {
				fmt.Println("== get token failed ===")
				common.FailedAuthReturn(c, "访问未授权")
				c.Abort()
				return
			}
		}
		res, err := userCahe.Value(token)
		if err != nil {
			common.FailedAuthReturn(c, "token已过期")
			c.Abort()
			return
		} else {
			// 验证通过，会继续访问下一个中间件
			user := res.Data().(User)
			if strings.HasPrefix(c.FullPath(), "/super") && strings.Compare(user.Role, ROLE_USER) == 0 {
				common.FailedAuthReturn(c, "用户无权访问该接口")
				c.Abort()
				return
			}
			c.Set("userInfo", user)
			c.Next()
		}
	}
}

func AdminCheck() gin.HandlerFunc {
	return func(context *gin.Context) {
		user, _ := context.Get("userInfo")
		var usr = user.(User)
		if strings.Compare(usr.Role, ROLE_ADMIN) == 0 {
			context.Next()
			return
		} else {
			common.FailedAuthReturn(context, "用户没有管理权限")
			context.Abort()
			return
		}
	}

}
