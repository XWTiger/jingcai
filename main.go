package main

import (
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	"jingcai/admin"
	"jingcai/advise"
	"jingcai/audit"
	"jingcai/bbs"
	"jingcai/claim"
	"jingcai/config"
	"jingcai/creeper"
	_ "jingcai/docs"
	"jingcai/files"
	ihttp "jingcai/http"
	alog "jingcai/log"
	"jingcai/mysql"
	"jingcai/order"
	"jingcai/score"
	"jingcai/shop"
	"jingcai/user"
	"os"
	"os/signal"
	"syscall"
)

var log = alog.Logger

func main() {
	//conf
	conf := config.Init()
	content, _ := yaml.Marshal(conf)
	log.Info(string(content))
	//start server
	clean := ihttp.Init(conf)
	files.Init(conf)

	//connect mysql
	sc := make(chan os.Signal, 1)
	if myErr := mysql.InitDB(conf); myErr != nil {
		log.Error("======== mysql connect failed =============")
		clean()
		os.Exit(1)
	}
	initTables()
	order.OrderCheckInit()
	ctx, cancel := context.WithCancel(context.Background())
	if conf.HttpConf.CreeperSwitch {
		go admin.InitCronForCreep(ctx)
	}
	//init audit log
	if conf.HttpConf.AuditSwitch {
		audit.InitAudit()
	}
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM) //syscall.SIGHUP
	for {
		sig := <-sc
		log.Info("收到信号：", sig.String())
		mysql.SqlDB.Close()
		break
	}
	cancel()
	clean()

}

func initTables() {
	log.Info("======== check mysql tables =============")
	mysql.DB.AutoMigrate(&order.AllWin{})
	mysql.DB.AutoMigrate(&order.Order{})
	mysql.DB.AutoMigrate(&order.Match{})
	mysql.DB.AutoMigrate(&user.Bill{})
	mysql.DB.AutoMigrate(&order.LotteryDetail{})
	mysql.DB.AutoMigrate(&order.Bet{})
	mysql.DB.AutoMigrate(&order.FootView{})
	mysql.DB.AutoMigrate(&files.FileStore{})
	mysql.DB.AutoMigrate(&order.OrderImage{})
	mysql.DB.AutoMigrate(&bbs.Comment{})
	mysql.DB.AutoMigrate(&bbs.Response{})
	mysql.DB.AutoMigrate(&creeper.Content{})
	mysql.DB.AutoMigrate(&shop.Shop{})
	mysql.DB.AutoMigrate(&user.User{})
	mysql.DB.AutoMigrate(&user.ScoreUserNotify{})
	mysql.DB.AutoMigrate(&advise.NotificationPO{})
	mysql.DB.AutoMigrate(&order.JobExecution{})
	mysql.DB.AutoMigrate(&audit.AuditLog{})
	mysql.DB.AutoMigrate(&claim.Claim{})
	mysql.DB.AutoMigrate(&score.FreeScore{})
}
