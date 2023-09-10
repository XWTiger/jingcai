package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	alog "jingcai/log"
)

var log = alog.Logger

type HttpConfig struct {
	Host            string `yaml:"Host"`
	Port            int    `yaml:"Port"`
	CertFile        string `yaml:"CertFile"`
	KeyFile         string `yaml:"KeyFile"`
	ShutdownTimeout int    `yaml:"ShutdownTimeout"`
	ReadTimeout     int    `yaml:"ReadTimeout"`
	WriteTimeout    int    `yaml:"WriteTimeout"`
	IdleTimeout     int    `yaml:"IdleTimeout"`
	TargetHost      string `yaml:"TargetHost"`
	TargetPort      int    `yaml:"TargetPort"`
	TargetScheme    string `yaml:"TargetScheme"`
	ImagePrefix     string `yaml:"ImagePrefix"`
	DbAddress       string `yaml:"DbAddress"`     //数据库
	DbUser          string `yaml:"DbUser"`        //用户
	DbSpecial       string `yaml:"DbSpecial"`     // 数据库密码
	CreeperSwitch   bool   `yaml:"CreeperSwitch"` // 爬虫开关
}

type Config struct {
	HttpConf *HttpConfig  `yaml:"Server"`
	LogConf  *alog.Config `yaml:"Log"`
}

func Init() *Config {
	file, err := ioutil.ReadFile("app.yaml")
	if err != nil {
		fmt.Println(err)
		log.Error("configuration read failed!! use default config")
		return &Config{
			HttpConf: &HttpConfig{
				Host:            "0.0.0.0",
				Port:            8888,
				ShutdownTimeout: 30,
				ReadTimeout:     20,
				WriteTimeout:    40,
				IdleTimeout:     120,
				TargetHost:      "127.0.0.1",
				TargetPort:      8888,
				TargetScheme:    "http",
				DbAddress:       "127.0.0.1",
				DbUser:          "root",
				DbSpecial:       "123456",
			},
			LogConf: &alog.Config{
				Path:        "/var/log/jingcai/jingcai.log",
				Debug:       false,
				KeepHours:   24 * 7,
				RotateHours: 24,
			},
		}
	}
	var conf Config
	err3 := yaml.Unmarshal(file, &conf)
	if err3 != nil {
		fmt.Println(err3)
		log.Error("yaml parse failed!! use default config")
		return &Config{
			HttpConf: &HttpConfig{
				Host:            "0.0.0.0",
				Port:            18080,
				ShutdownTimeout: 30,
				ReadTimeout:     20,
				WriteTimeout:    40,
				IdleTimeout:     120,
				TargetHost:      "127.0.0.1",
				TargetPort:      8080,
				TargetScheme:    "http",
				CreeperSwitch:   true,
			},
			LogConf: &alog.Config{
				Path:        "/var/log/jingcai/jingcai.log",
				Debug:       false,
				KeepHours:   24 * 7,
				RotateHours: 24,
			},
		}
	}
	return &conf
}
