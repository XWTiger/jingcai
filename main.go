package main

import (
	/*_ "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"*/
	"gopkg.in/yaml.v3"
	"jingcai/config"
	_ "jingcai/docs"
	ihttp "jingcai/http"
	alog "jingcai/log"
	"jingcai/mysql"
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
	//connect mysql
	sc := make(chan os.Signal, 1)
	if myErr := mysql.InitDB(); myErr != nil {
		log.Error("======== mysql connect failed =============")
		clean()
		os.Exit(1)
	}

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		sig := <-sc
		log.Info("收到信号：", sig.String())
		mysql.SqlDB.Close()
		break
	}
	clean()

}
