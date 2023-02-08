package log

import (
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"time"
)

var Logger = logrus.New()

type Config struct {
	Path        string `yaml:"Path"`
	Debug       bool   `yaml:"Debug"`
	KeepHours   uint   `yaml:"KeepHours"`
	RotateHours uint   `yaml:"RotateHours"`
}

func Init(c *Config) {
	writer, _ := rotatelogs.New(c.Path+".%Y-%m-%d",
		rotatelogs.WithLinkName(c.Path),
		rotatelogs.WithMaxAge(time.Duration(c.KeepHours)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(c.RotateHours)*time.Hour),
	)
	// 设置日志输出
	Logger.SetOutput(writer)
	// 设置日志输出格式，代替默认的ASCII格式
	Logger.SetFormatter(&logrus.TextFormatter{})
	Logger.SetReportCaller(true)
	// 设置日志等级
	if c.Debug {
		Logger.SetLevel(logrus.DebugLevel)
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}
