package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"jingcai/admin"
	"jingcai/bbs"
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
	s := r.Group("/super")
	{
		s.GET("/creep", admin.CreepHandler)
	}

	bbsGroup := r.Group("/bbs")
	{
		bbsGroup.GET("/list", bbs.ListHandler)
		bbsGroup.POST("/commit", bbs.CommitHandler)
	}

}

// @Description 状态检测
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /ping [get]
func pong(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
