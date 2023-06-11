package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"jingcai/admin"
	"jingcai/advise"
	"jingcai/bbs"
	"jingcai/cache"
	"jingcai/common"
	"jingcai/files"
	"jingcai/mysql"
	"jingcai/order"
	"jingcai/user"
	"net/http"
)

// @title           黑马推荐接口
// @version         1.0
// @description     推荐足球，篮球等相关信息

// @contact.name   tiger
// @contact.url    http://www.swagger.io/support
// @contact.email  tiger

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth
func BindRouters(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/ping", pong)
	r.GET("/salt", common.Salt)
	r.GET("/notify", advise.Query)
	r.GET("/tiger-dragon-list", order.TigerDragonList)
	r.POST("/cache", cache.Set)
	userGroup := r.Group("/user")
	{
		userGroup.POST("", user.UserCreateHandler)
		userGroup.POST("/login", user.Login)
		userGroup.POST("/logout", user.Logout)
	}

	r.Use(user.Authorize())
	bbsGroup := r.Group("/bbs")
	r.POST("/user/complain", user.UserComplain)
	r.GET("/user/info", user.GetUserInfo)
	r.POST("/user/info", user.UpdateUser)
	s := r.Group("/super")
	{
		s.GET("/creep", admin.CreepHandler)
		s.GET("/complains", admin.ListComplain)
		s.POST("/notify", advise.Create)
	}
	{
		bbsGroup.GET("/list", bbs.ListHandler)
		bbsGroup.POST("/commit", bbs.CommitHandler)
	}
	//订单
	orderGroup := r.Group("/order")
	{
		orderGroup.POST("", order.OrderCreate)
		orderGroup.GET("", order.OrderList)
		orderGroup.GET("/bets", order.GetBetByOrder)
		orderGroup.GET("/all_win", order.AllWinList)
		orderGroup.POST("/all_win", order.AllWinCreateHandler)
		orderGroup.POST("/follow", order.FollowOrder)
	}

	//文件上传下载
	r.POST("/upload", files.Upload)
	r.GET("/download", files.DownLoad)

}

// @Description 状态检测
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /ping [get]
func pong(c *gin.Context) {
	pwd, _ := common.DePwdCode("rBhpl45Z3NpBxYhMuAuIqA==", []byte("c5b55acf-b0d4-43"))
	fmt.Println(string(pwd))
	var user2 user.User
	mysql.DB.Model(&user.User{Model: gorm.Model{
		ID: 1,
	}}).Find(&user2)
	fmt.Println(user2.Name)

	c.String(http.StatusOK, "pong")

}
