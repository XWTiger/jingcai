package main

import (
	"gopkg.in/yaml.v3"
	"jingcai/config"
	_ "jingcai/docs"
	ihttp "jingcai/http"
	alog "jingcai/log"
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
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		sig := <-sc
		log.Info("收到信号：", sig.String())
		break
	}
	clean()

}
