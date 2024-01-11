package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"jingcai/admin"
	"jingcai/advise"
	"jingcai/audit"

	"jingcai/common"
	"jingcai/config"
	"jingcai/files"

	"jingcai/order"
	"jingcai/shop"
	"jingcai/user"
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
func BindRouters(g *gin.Engine, config *config.Config) {

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	g.GET("/ping", pong)
	r := g.Group("/api")
	if config.HttpConf.AuditSwitch {
		if config.HttpConf.AuditSwitch {
			r.Use(audit.AuditHandler())
		}
	}
	shopGroup := r.Group("/shop")
	userGroup := r.Group("/user")
	r.GET("/salt", common.Salt)
	r.GET("/notify", advise.Query)
	r.GET("/tiger-dragon-list", order.TigerDragonList)

	//广告类接口
	//adGroup := r.Group("/advertising")

	//r.GET("/ws", websocket.OrderWebSocket)
	{
		userGroup.POST("", user.UserCreateHandler)
		userGroup.POST("/login", user.Login)
		userGroup.POST("/logout", user.Logout)
		userGroup.POST("/passwordByPhoneCode", user.ChangePasswordByPhoneCodeHandler)
	}

	r.GET("/download/:name", files.DownLoad)

	r.POST("/shop", shop.ShopRegistry)
	r.Use(user.Authorize())
	{
		userGroup.Use(user.Authorize())
		userGroup.POST("/bill/notify", user.BillClearNotify)
		userGroup.GET("/bill/notify", user.BillClearNotifyList)
		userGroup.GET("/owner", user.GetShopOwnerInfo)
		userGroup.POST("/password", user.ChangePasswordHandler)
	}
	r.POST("/user/complain", user.UserComplain)
	r.GET("/user/info", user.GetUserInfo)
	r.POST("/user/info", user.UpdateUser)
	s := r.Group("/super")
	{
		s.Use(user.AdminCheck())
		s.GET("/complains", admin.ListComplain)
		s.POST("/notify", advise.Create)

		s.POST("/add-score", user.AddScore)

		s.GET("/statistics", shop.StatisticsCount)
		s.POST("/substract-score", user.BillClear)
		s.GET("/bills", shop.ShopBills)
		s.GET("/bill/notify", user.BillClearShopNotifyList)
	}

	{
		//店铺注册
		shopGroup.GET("/users", shop.QueryShopUser)

	}

	//订单
	orderGroup := r.Group("/order")
	{
		orderGroup.POST("", order.OrderCreate)

	}

	//文件上传下载
	r.POST("/upload", files.Upload)

}

// @Description 状态检测
// @Summery 状态检测
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /ping [get]
func pong(c *gin.Context) {
	fmt.Println(fmt.Sprintf("%d", 42))

	/*var obj map[string]interface{}
	cont, _ := swag.ReadDoc("swagger")
	json.Unmarshal([]byte(cont), &obj)
	fmt.Println("===========================")
	val := obj["paths"].(map[string]interface{})
	for k, v := range val {
		fmt.Println(k)
		valin := v.(map[string]interface{})
		if valin["get"] != nil {
			detail := valin["get"].(map[string]interface{})
			fmt.Println(detail["description"])

		}
	}*/

}
