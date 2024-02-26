package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"jingcai/admin"
	"jingcai/advise"
	"jingcai/audit"
	"jingcai/bbs"
	"jingcai/cache"
	"jingcai/claim"
	"jingcai/common"
	"jingcai/config"
	"jingcai/files"
	"jingcai/lottery"
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
	r.GET("/dict", common.Dict)
	r.GET("/notify", advise.Query)
	r.GET("/tiger-dragon-list", order.TigerDragonList)
	r.POST("/cache", cache.Set)
	lott := g.Group("/lottery-api")
	{
		lott.GET("/seven-star", lottery.SevenStarFun)
		lott.GET("/plw", lottery.PlwFun)
		lott.GET("/super-lottery", lottery.SuperLotteryFun)
		lott.GET("/statistics", lottery.LotteryStatisticsHandler)
	}

	//广告类接口
	adGroup := r.Group("/advertising")
	{
		adGroup.GET("/win-list", order.GetWinUserList)
	}

	//r.GET("/ws", websocket.OrderWebSocket)
	{
		userGroup.POST("", user.UserCreateHandler)
		userGroup.POST("/login", user.Login)
		userGroup.POST("/passwordByPhoneCode", user.ChangePasswordByPhoneCodeHandler)
	}

	r.GET("/download/:name", files.DownLoad)
	bbsGroup := r.Group("/bbs")

	r.POST("/shop", shop.ShopRegistry)
	r.Use(user.Authorize())
	{
		userGroup.Use(user.Authorize())
		userGroup.POST("/logout", user.Logout)
		userGroup.POST("/bill/notify", user.BillClearNotify)
		userGroup.GET("/bill/notify", user.BillClearNotifyList)
		userGroup.GET("/owner", user.GetShopOwnerInfo)
		userGroup.POST("/password", user.ChangePasswordHandler)
		userGroup.POST("/claims", claim.UserClaim)
	}
	r.POST("/user/complain", user.UserComplain)
	r.GET("/user/info", user.GetUserInfo)
	r.POST("/user/info", user.UpdateUser)
	s := r.Group("/super")
	{
		s.Use(user.AdminCheck())
		s.GET("/complains", admin.ListComplain)
		s.POST("/notify", advise.Create)
		s.POST("/check/lottery_check", order.AddCheckForManual)
		s.POST("/add-score", user.AddScore)
		s.GET("/order", order.AdminOrderList)
		s.POST("/bets", order.UploadBets)
		s.GET("/statistics", shop.StatisticsCount)
		s.POST("/substract-score", user.BillClear)
		s.GET("/bills", shop.ShopBills)
		s.GET("/bill/notify", user.BillClearShopNotifyList)
		s.GET("/claims", claim.ClaimList)
		s.POST("/claims", claim.UpdateClaim)
		s.POST("/matches/odds", order.UpdateOddHandler)
	}

	{
		//店铺注册
		shopGroup.GET("/users", shop.QueryShopUser)

	}

	{
		bbsGroup.GET("/comment/list", bbs.ListComment)
		bbsGroup.GET("/list", bbs.ListHandler)
		bbsGroup.GET("/comment/response", bbs.GetResponseByCommentId)
		bbsGroup.Use(user.Authorize())
		bbsGroup.POST("/commit", bbs.CommitHandler)
		bbsGroup.POST("/response", bbs.ResponseHandler)
		bbsGroup.POST("/comment", bbs.CommentHandler)
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
		orderGroup.GET("/shared", order.SharedOrderList)
	}
	adminGroup := r.Group("/admin")
	{
		adminGroup.GET("/creep", admin.CreepHandler)

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
	//fmt.Println(order.GetOrderId(&order.Order{UserID: 1, LotteryType: "FOOTBALL", Share: true}))
	var tmp = [][]string{{"1", "2"}, {"3", "4"}, {"5", "6"}, {"7", "8"}, {"9", "10"}, {"11"}, {"12", "13"}}
	var strace = []string{}
	var res = [][]string{}
	order.GetIndexCmn(0, tmp, &strace, &res)
	for _, re := range res {
		fmt.Println(re)
	}
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
