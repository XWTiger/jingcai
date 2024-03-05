package admin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"golang.org/x/net/context"
	"jingcai/config"
	"jingcai/creeper"
	"jingcai/mysql"
	"math/rand"
	"net/http"
	"time"
)
import ilog "jingcai/log"

var log = ilog.Logger
var creeperSwitch = true

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

func Init(conf config.Config) {
	creeperSwitch = conf.HttpConf.CreeperSwitch
}

func (cc *CreepCenter) Doing() error {
	initTables()
	tx := mysql.DB.Begin()
	for _, c := range cc.Creepers {
		//log.Info("url=", k)
		content := c.Creep()

		for i, _ := range content {
			content[i].FilterStrInSummery()
		}

		tx.Create(&content)
		for _, ctx := range content {
			//log.Info("content: ", ctx.Content, "url: ", ctx.Url)
			if len(ctx.Conditions) > 0 {
				for _, cond := range ctx.Conditions {
					tx.Create(&creeper.Condition{
						ParentId:  ctx.ID,
						Condition: cond,
					})
				}
			}
		}
	}
	tx.Commit()
	return nil
}

// @Summary 爬虫接口
// @Description 爬虫接口
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Router /api/admin/creep [get]
func CreepHandler(c *gin.Context) {

	if !creeperSwitch {
		log.Warn("======= 该服务并没有开启爬虫功能 ============")
		return
	}
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(100)
	time.Sleep(time.Second * time.Duration(num))
	tianIns := creeper.NewTianInstance()
	CreepRegistry.registry(tianIns)
	Leisu := creeper.NewInstance()
	CreepRegistry.registry(Leisu)
	CreepRegistry.Doing()
	if c != nil {
		c.String(http.StatusOK, "finished creep")
	} else {
		log.Info("====== 爬虫定时任务结束 =======")
	}
}

func InitCronForCreep(ctx context.Context) {
	c := cron.New()
	spec := "30 */10 * * *"
	err := c.AddFunc(spec, func() {
		CreepHandler(nil)
	})
	fmt.Println(err)
	c.Start()

	select {
	case <-ctx.Done():
		fmt.Println("======== 爬虫定时任务退出 ========")
		return

	}
}
