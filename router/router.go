package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"jingcai/admin"
	"jingcai/advise"
	"jingcai/bbs"
	"jingcai/cache"
	"jingcai/common"
	"jingcai/files"
	"jingcai/lottery"
	"jingcai/order"
	"jingcai/shop"
	"jingcai/user"
	"jingcai/util"
	"strings"
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
func BindRouters(g *gin.Engine) {

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	g.GET("/ping", pong)
	r := g.Group("/api")
	shopGroup := r.Group("/shop")
	userGroup := r.Group("/user")
	r.GET("/salt", common.Salt)
	r.GET("/notify", advise.Query)
	r.GET("/tiger-dragon-list", order.TigerDragonList)
	r.POST("/cache", cache.Set)
	lott := g.Group("/lottery-api")
	{
		lott.GET("/seven-star", lottery.SevenStarFun)
		lott.GET("/plw", lottery.PlwFun)
		lott.GET("/super-lottery", lottery.SuperLotteryFun)
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
		userGroup.POST("/logout", user.Logout)
	}

	r.GET("/download/:name", files.DownLoad)
	bbsGroup := r.Group("/bbs")

	r.POST("/shop", shop.ShopRegistry)
	r.Use(user.Authorize())
	{
		userGroup.Use(user.Authorize())
		userGroup.POST("/bill/notify", user.BillClearNotify)
		userGroup.GET("/bill/notify", user.BillClearNotifyList)
		userGroup.GET("/owner", user.GetShopOwnerInfo)
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
	}

	{
		//店铺注册
		shopGroup.GET("/users", shop.QueryShopUser)

	}

	{
		bbsGroup.POST("/commit", bbs.CommitHandler)
		bbsGroup.GET("/list", bbs.ListHandler)
		bbsGroup.POST("/response", bbs.ResponseHandler)
		bbsGroup.POST("/comment", bbs.CommentHandler)
		bbsGroup.GET("/comment/list", bbs.ListComment)
		bbsGroup.GET("/comment/response", bbs.GetResponseByCommentId)
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
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /ping [get]
func pong(c *gin.Context) {
	/*num, _ := strconv.Atoi("01")
	fmt.Println(num)
	pwd, _ := common.DePwdCode("rBhpl45Z3NpBxYhMuAuIqA==", []byte("c5b55acf-b0d4-43"))
	fmt.Println(string(pwd))
	var user2 user.User
	mysql.DB.Model(&user.User{Model: gorm.Model{
		ID: 1,
	}}).Find(&user2)
	fmt.Println(user2.Name)
	*/
	/*arr := []int{1, 2, 3}
	result := util.Permute(arr)
	for _, ints := range result {
		fmt.Println(ints)
	}*/
	/*arr := []int{1, 2, 2}
	arr2 := []int{3, 1, 3}
	arr3 := []int{4, 4, 1}
	fmt.Println(util.GetCombine3(arr))
	fmt.Println(util.GetCombine3(arr2))
	fmt.Println(util.GetCombine3(arr3))
	c.String(http.StatusOK, "pong")*/
	/*fmt.Println(order.GetOrderId(&order.Order{

		LotteryType: order.P3,
		Share:       true,
	}))*/
	/*var all = [][]string{{"0", "1"}, {"0", "1"}, {"0", "1"}}
	var childs = make([]string, 0)
	var sb = make([]byte, 0)
	util.GetZxGsb(0, all, &sb, &childs)
	fmt.Println(childs[0])*/

	var all = []int{7, 7, 4, 1, 2, 3}
	result := util.PermuteAnm(all, 5)
	var sum = 0

	var duplicate = make([][]int, 0)
	for _, ints := range result {
		var exist = false
		for _, value := range duplicate {
			tmp := fmt.Sprintf("%d%d%d%d%d", value[0], value[1], value[2], value[3], value[4])
			tmp2 := fmt.Sprintf("%d%d%d%d%d", ints[0], ints[1], ints[2], ints[3], ints[4])
			//fmt.Println(tmp2)
			if strings.Compare(tmp2, tmp) == 0 {
				exist = true
				break
			}
		}
		if !exist {
			duplicate = append(duplicate, ints)
		}
	}
	fmt.Println(len(duplicate))
	for _, ints := range result {
		var count = 0
		for _, val := range ints {
			if val == 7 {
				count++
				if count == 2 {
					break
				}
			}

		}
		if count == 2 {

			sum++
			fmt.Println(ints)
		}
	}
	fmt.Println(sum)
}
