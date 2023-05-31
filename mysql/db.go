package mysql

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	ilog "jingcai/log"
)

var DB *gorm.DB
var SqlDB *sql.DB
var log = ilog.Logger

func InitDB() error {
	dbIns, err := gorm.Open(mysql.Open("root:123456@(127.0.0.1:3306)/jingcai?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		log.Error("mysql connect failed", err)
		return err
	}
	sqlDB, _ := dbIns.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(2048)
	DB = dbIns
	SqlDB = sqlDB

	return nil
}
