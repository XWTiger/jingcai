package admin

import (
	"github.com/gin-gonic/gin"
	"jingcai/creeper"
	"jingcai/mysql"
	"net/http"
)
import ilog "jingcai/log"

var log = ilog.Logger

type CreepCenterInterface interface {
	registry() error
	Doing() error
}
type CreepCenter struct {
	Creepers map[string]creeper.Creeper
}

var CreepRegistry = &CreepCenter{
	Creepers: make(map[string]creeper.Creeper),
}

func (cc *CreepCenter) registry(member creeper.Creeper) error {
	if cc.Creepers[member.Key()] == nil {
		cc.Creepers[member.Key()] = member
	}
	log.Info(len(cc.Creepers))
	return nil
}
func initTables() {
	mysql.DB.AutoMigrate(&creeper.Content{})
	mysql.DB.AutoMigrate(&creeper.Condition{})
}
func (cc *CreepCenter) Doing() error {
	initTables()
	for k, c := range cc.Creepers {
		log.Info("url=", k)
		content := c.Creep()
		mysql.DB.Create(&content)
		for _, ctx := range content {
			log.Info("content: ", ctx.Content, "url: ", ctx.Url)
			if len(ctx.Conditions) > 0 {
				for _, cond := range ctx.Conditions {
					mysql.DB.Create(&creeper.Condition{
						ParentId:  ctx.ID,
						Condition: cond,
					})
				}
			}
		}
	}
	return nil
}

// @Description 爬虫接口
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /super/creep [get]
func CreepHandler(c *gin.Context) {
	tianIns := creeper.NewTianInstance()
	CreepRegistry.registry(tianIns)
	Leisu := creeper.NewInstance()
	CreepRegistry.registry(Leisu)
	CreepRegistry.Doing()
	c.String(http.StatusOK, "finished creep")
}
