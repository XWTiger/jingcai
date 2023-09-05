package mysql

import (
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"jingcai/config"
	ilog "jingcai/log"
)

var DB *gorm.DB
var SqlDB *sql.DB
var log = ilog.Logger

func InitDB(conf *config.Config) error {
	connUrl := fmt.Sprintf("%s:%s@(%s:3306)/jingcai?charset=utf8mb4&parseTime=True&loc=Local", conf.HttpConf.DbUser, conf.HttpConf.DbSpecial, conf.HttpConf.DbAddress)
	fmt.Println(connUrl)
	dbIns, err := gorm.Open(mysql.Open(connUrl), &gorm.Config{})
	if err != nil {
		log.Error("mysql connect failed", err)
		return err
	}
	dbIns.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8")
	sqlDB, _ := dbIns.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(2048)
	DB = dbIns
	SqlDB = sqlDB

	return nil
}
