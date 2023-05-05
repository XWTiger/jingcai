package mysql

import (
	"github.com/jinzhu/gorm"
	ilog "jingcai/log"
)

var db gorm.DB
var log = ilog.Logger

func initDB() {
	db, err := gorm.Open("mysql", "root:root@(127.0.0.1:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Error("mysql connect failed", err)
	}
}
