package main

import (
	_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v3"
	"jingcai/admin"
	"jingcai/bbs"
	"jingcai/config"
	"jingcai/creeper"
	_ "jingcai/docs"
	"jingcai/files"
	ihttp "jingcai/http"
	alog "jingcai/log"
	"jingcai/mysql"
	"jingcai/order"
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
	ctx, cancel := context.WithCancel(context.Background())
	go admin.InitCronForCreep(ctx)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
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
	mysql.DB.AutoMigrate(&order.LotteryDetail{})
	mysql.DB.AutoMigrate(&order.Bet{})
	mysql.DB.AutoMigrate(&order.FootView{})
	mysql.DB.AutoMigrate(&files.FileStore{})
	mysql.DB.AutoMigrate(&order.OrderImage{})
	mysql.DB.AutoMigrate(&bbs.Comment{})
	mysql.DB.AutoMigrate(&bbs.Response{})
	mysql.DB.AutoMigrate(&creeper.Content{})
}
